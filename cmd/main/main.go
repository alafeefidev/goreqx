package main

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
			"sus": "yeah",
		},
		JSON: map[string]any{
			"value": 5,
			"worth": nil,
		},
		Headers: map[string]string{
			"X-CacheZ": "Noice",
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
		cachez.WithRandomUserAgent(),
	)
	cache.StartEviction(ctx, 1*time.Minute)

	var wg sync.WaitGroup

	start := time.Now()
	for range 5 {
		wg.Add(1)
		go req(cache, ctx, &wg)
	}
	wg.Wait()
	fmt.Println("All done")
	fmt.Println(time.Since(start))

}
