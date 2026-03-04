package goreqx

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
	URL    string
	Method string
}

type Cache struct {
	mu               sync.RWMutex
	entries          map[Request]*CachedResponse
	client           *http.Client
	logger           *slog.Logger
	skipOnBadRequest bool
}

type Option func(*Cache)

func WithLogger(sl *slog.Logger) Option {
	return func(c *Cache) { c.logger = sl }
}

func WithClient(cl *http.Client) Option {
	return func(c *Cache) { c.client = cl }
}

// Create a new instance of Cache
func New(opts ...Option) *Cache {
	c := &Cache{
		entries:          make(map[Request]*CachedResponse),
		skipOnBadRequest: true,
		logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
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
	entry, ok := c.entries[req]
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
	c.entries[req] = entry
	c.mu.Unlock()
}

func (c *Cache) delete(req Request) {
	c.mu.Lock()
	delete(c.entries, req)
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

func (c *Cache) Do(rq Request, ttl time.Duration) (*CachedResponse, error) {
	if entry, ok := c.get(rq); ok {
		c.logger.Info("cache hit", "url", rq.URL)
		return entry, nil
	}

	c.logger.Info("cache miss", "url", rq.URL)

	req, err := http.NewRequest(rq.Method, rq.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	entry, err := CacheFromResponse(resp, ttl)
	if err != nil {
		return nil, err
	}

	if c.skipOnBadRequest && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		c.logger.Warn("Skipping caching", "url", rq.URL, "reason", fmt.Sprintf("%d status code", resp.StatusCode))
	} else {
		c.set(rq, entry)
	}

	return entry, nil
}

func (c *Cache) Get(url string, ttl time.Duration) (*CachedResponse, error) {
	req := Request{
		URL:    url,
		Method: "GET",
	}

	entry, err := c.Do(req, ttl)
	if err != nil {
		return nil, err
	}
	return entry, nil
}
