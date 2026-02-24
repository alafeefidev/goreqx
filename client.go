package goreqx

import (
	"net/http"
	"time"
)

type reqxClient struct {
	AClient http.Client
	MaxAge time.Time
	// more settings
}

type Response struct {
	Code         int `json:"code"` // Status code, 200...
	ResponseType string `json:"response_type"` // JSON and HTML
	HBody        string `json:"hbody"`
	JBody        byte `json:"jbody"`
}

var DefaultReqxClient = &reqxClient{}

func (r *reqxClient) DoBody(req *http.Request) *Response {
	
}


// func main() {
// 	http.NewRequestWithContext()
// 	http.DefaultClient.Do()
// 	DefaultReqxClient.AClient.Do()
// }