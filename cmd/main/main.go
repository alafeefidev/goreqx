package main

import (
	"github.com/alafeefidev/goreqx"
	"net/http"
)

func main() {
	c := goreqx.DefaultConfig
	
	req, err := http.NewRequest(http.MethodGet, "https://httpbin.org/get", nil)
	

}
