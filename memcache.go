package cachez

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Memory-Based Cache

type CachedResponse struct {
	Body       []byte
	Headers    http.Header
	Cookies    []*http.Cookie
	StatusCode int
	CachedAt   time.Time
	TTL        time.Duration
}

type Request struct {
	// Needed
	URL    string
	Method string

	// Optional
	Headers map[string]string
	Cookies []*http.Cookie
	Query   map[string]string

	// Optional, choose one
	Form map[string]string
	JSON map[string]any
}

type Cache struct {
	mu          sync.RWMutex
	entries     map[string]*CachedResponse
	flight      singleflight.Group
	client      *http.Client
	logger      *slog.Logger
	shouldCache func(StatusCode int) bool
	userAgent   string
}

type Option func(*Cache)

// This function solves that I can't pass Request directly as a key to the entries key in the Cache struct
// because a struct of non-comparable types is not allowed to be used as a map key.
// Here we're converting the Request struct to a string that is comparable and deterministic.
func (r Request) ToKey() string {
	refactored := struct {
		URL     string
		Method  string
		Headers [][2]string
		Cookies [][2]string
		Query   [][2]string
		Form    [][2]string
		JSON    []byte
	}{
		URL:     r.URL,
		Method:  r.Method,
		Headers: mapSlice(r.Headers),
		Cookies: cookieSlice(r.Cookies),
		Query:   mapSlice(r.Query),
		Form:    mapSlice(r.Form),
		JSON:    formatJSON(r.JSON),
	}
	data, _ := json.Marshal(refactored)
	hash := sha256.Sum256(data)
	// hex hash
	return fmt.Sprintf("%x", hash)

}

func WithLogger(sl *slog.Logger) Option {
	return func(c *Cache) { c.logger = sl }
}

func WithClient(cl *http.Client) Option {
	return func(c *Cache) { c.client = cl }
}

func WithCacheSuccessOnly() Option {
	return func(c *Cache) {
		c.shouldCache = func(code int) bool {
			return code >= 200 && code < 300
		}
	}
}

func WithCacheAll() Option {
	return func(c *Cache) {
		c.shouldCache = func(code int) bool { return true }
	}
}

func WithCacheCustom(fn func(StatusCode int) bool) Option {
	return func(c *Cache) { c.shouldCache = fn }
}

func WithRandomUserAgent() Option {
	return func(c *Cache) {
		c.userAgent = RandomUserAgent()
	}
}

func WithCustomUserAgent(userAgent string) Option {
	return func(c *Cache) { c.userAgent = userAgent }
}

// Create a new instance of Cache
func New(opts ...Option) *Cache {
	c := &Cache{
		entries: make(map[string]*CachedResponse),
		logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		client:  http.DefaultClient,
		shouldCache: func(code int) bool {
			return code >= 200 && code < 300
		},
		userAgent: "Go-http-client/2.0",
	}

	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *CachedResponse) IsExpired() bool {
	return time.Since(c.CachedAt) > c.TTL
}

func (c *Cache) get(req Request) (*CachedResponse, bool) {
	c.mu.RLock()
	entry, ok := c.entries[req.ToKey()]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if entry.IsExpired() {
		c.delete(req)
		return nil, false
	}

	return entry, true
}

func (c *Cache) set(req Request, entry *CachedResponse) {
	c.mu.Lock()
	c.entries[req.ToKey()] = entry
	c.mu.Unlock()
}

func (c *Cache) delete(req Request) {
	c.mu.Lock()
	delete(c.entries, req.ToKey())
	c.mu.Unlock()
}

func CacheFromResponse(resp *http.Response, ttl time.Duration) (*CachedResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &CachedResponse{
		Body:       body,
		Headers:    resp.Header.Clone(),
		Cookies:    resp.Cookies(),
		StatusCode: resp.StatusCode,
		CachedAt:   time.Now(),
		TTL:        ttl,
	}, nil
}

func (c *Cache) Do(ctx context.Context, rq Request, ttl time.Duration) (*CachedResponse, error) {

	if rq.URL == "" || rq.Method == "" {
		return nil, errors.New("no url or method provided")
	}

	if entry, ok := c.get(rq); ok {
		c.logger.Info("cache hit", "url", rq.URL)
		return entry, nil
	}

	// Make sure if multiple goroutines are requesting the same url then only one do the actual work
	// the rest get the first result, only happens at the same time
	val, err, _ := c.flight.Do(rq.ToKey(), func() (interface{}, error) {
		c.logger.Info("cache miss", "url", rq.URL)

		u, err := url.Parse(rq.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid url: %w", err)
		}
		q := u.Query()
		for k, v := range rq.Query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()

		var body io.Reader
		var contentType string

		switch {
		case rq.JSON != nil:
			data, err := json.Marshal(rq.JSON)
			if err != nil {
				return nil, fmt.Errorf("error in json marshal: %w", err)
			}
			body = bytes.NewReader(data)
			contentType = "application/json"
		case rq.Form != nil:
			form := url.Values{}
			for k, v := range rq.Form {
				form.Set(k, v)
			}
			body = strings.NewReader(form.Encode())
			contentType = "application/x-www-form-urlencoded"
			}
		

		req, err := http.NewRequestWithContext(ctx, rq.Method, u.String(), body)
		if err != nil {
			return nil, err
		}

		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		for k, v := range rq.Headers {
			req.Header.Set(k, v)
		}

		// Maybe remove the user agent from replacing user defined
		req.Header.Set("User-Agent", c.userAgent)

		for _, c := range rq.Cookies {
			req.AddCookie(c)
		}

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		entry, err := CacheFromResponse(resp, ttl)
		if err != nil {
			return nil, err
		}

		if c.shouldCache(resp.StatusCode) {
			c.set(rq, entry)
		} else {
			c.logger.Warn("Skipping caching", "url", rq.URL, "status code", resp.StatusCode)
		}

		return entry, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(*CachedResponse), nil

}

func (c *Cache) StartEviction(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.mu.Lock()
				for k, v := range c.entries {
					if v.IsExpired() {
						delete(c.entries, k)
					}
				}
				c.mu.Unlock()
			}
		}
	}()

}

func (c *Cache) Get(ctx context.Context, url string, ttl time.Duration) (*CachedResponse, error) {
	req := Request{
		URL:    url,
		Method: http.MethodGet,
	}

	entry, err := c.Do(ctx, req, ttl)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (c *Cache) Post(ctx context.Context, url string, ttl time.Duration) (*CachedResponse, error) {
	req := Request{
		URL:    url,
		Method: http.MethodPost,
	}

	entry, err := c.Do(ctx, req, ttl)
	if err != nil {
		return nil, err
	}
	return entry, nil
}
