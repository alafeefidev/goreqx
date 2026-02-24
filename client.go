package goreqx

import (
	"fmt"
	"net/http"
)

var DefaultReqxClient = &reqxClient{}

type reqxClient struct {
	AClient http.Client
	// more settings
}

func (r *reqxClient) DoBody(req *http.Request)


func main() {
	http.NewRequestWithContext()
	http.DefaultClient.Do()
	DefaultReqxClient.AClient.Do()
}