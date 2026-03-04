# CacheZ

A go library for use in development of any service requiring http requests, It allows caching of responses in memory to not overstress external endpoints or websites in general.


## Installation

```go
import "github.com/alafeefidev/cachez"
```

## Features
- Full caching of requests for a specified amount of time
- Extensive options and settings for custom requests
- Goroutines safe

## Planned Features
- File-based caching
- DNS caching
- More options

## Usage
### Simple
```go
import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alafeefidev/cachez"
)


func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize a new Cache instance with logging enabled, and randomize user agents
	cache := cachez.New(
		cachez.WithLogger(slog.Default()),
		cachez.WithRandomUserAgent(),
	)

	// StartEviction starts a background goroutine to clean expired cache
	cache.StartEviction(ctx, 1*time.Minute)
	
	// Doing a GET request using the context, url, and the TTL (Time To Live)
	// It caches the output automatically
	resp, _ := cache.Get(ctx, "https://httpbin.org/get", 5*time.Minute)

	fmt.Println(string(resp.Body)) // Cache miss

	resp2, _ := cache.Get(ctx, "https://httpbin.org/get", 5*time.Minute)

	fmt.Println(string(resp2.Body)) // Cache hit
}
```
### Advanced
```go
import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/alafeefidev/cachez"
)

func req(cache *cachez.Cache, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	rq := cachez.Request{
		URL: "https://httpbin.org/post",
		Method: http.MethodPost,
		Query: map[string]string{
			"name": "yo",
			"see": "yeah",
		},
		JSON: map[string]any{
			"value": 5,
			"worth": nil,
		},
		Headers: map[string]string{
			"X-CacheZ": "Nice",
		},
	}

	resp, err := cache.Do(ctx, rq, 5*time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(resp.Body))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()


	cache := cachez.New(
		cachez.WithLogger(slog.Default()),
		cachez.WithCacheAll(),
		cachez.WithCustomUserAgent("cachez"),
	)
	cache.StartEviction(ctx, 5*time.Minute)
	
	var wg sync.WaitGroup

	for range 5 {
		wg.Add(1)
		go req(cache, ctx, &wg)
	}
	wg.Wait()
	fmt.Println("All done")
}
```
