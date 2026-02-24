package goreqx

import (
	"time"
	"github.com/google/uuid"
)

type ReqCache struct {
	// Check he url as a url not string
	Id uuid.UUID `json:"id"`
	Code int `json:"code"` // Status code, 200...
	Url string `json:"url"`
	FileType string `json:"file_type"` // JSON and HTML
}

type Config struct {
	LastModified time.Time `json:"last_modified"`
	Caches []ReqCache `json:"caches"`
}

var DefaultClient = &Config{}