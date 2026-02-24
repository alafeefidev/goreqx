package goreqx

import (
	"time"
	"github.com/google/uuid"
)

type ReqCache struct {
	// Check he url as a url not string
	Url string `json:"url"`
	Id uuid.UUID

}

type Config struct {
	LastModified time.Time `json:"last_modified"`
	Caches []ReqCache `json:"caches"`
}

var DefaultClient = &Config{}