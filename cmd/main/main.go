package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/alafeefidev/goreqx"
)

//TODO memory and file based caching

func main() {
	url := "https://httpbin.org/status/490"
	url2 := "https://httpbin.org/status/490"
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cache := goreqx.New(
		goreqx.WithLogger(logger),
	)

	cache.Get(url, 25*time.Second)
	cache.Get(url2, 25*time.Second)
	cache.Get(url, 25*time.Second)
	cache.Get(url2, 25*time.Second)
	cache.Get(url, 25*time.Second)
	cache.Get(url2, 25*time.Second)
	cache.Get(url, 25*time.Second)

}
