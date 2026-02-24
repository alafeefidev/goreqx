package goreqx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

var (
	configFolder string = "goreqx_cache"
	configFile   string = filepath.Join(configFolder, "goreqx.json")
)

type ReqCache struct {
	// Check he url as a url not string
	Id           uuid.UUID `json:"id"`
	Code         int       `json:"code"` // Status code, 200...
	Url          string    `json:"url"`
	FileType     string    `json:"file_type"` // JSON and HTML
	LastModified time.Time  `json:"last_modified"`
}

type Config struct {
	LastModified time.Time  `json:"last_modified"`
	Caches       []ReqCache `json:"caches"`
}

var DefaultConfig = &Config{
	LastModified: time.Now(),
	Caches:       make([]ReqCache, 0),
}

func AddConfig(resp *Response) error {
	InitConfig()


}

func GetCacheFile(*ReqCache) (any, error)
//TODO fix to return json or html

func GetConfig(url string) (*ReqCache, error) {
	exist, err := InitConfig()
	if err != nil {
		return nil, err
	} else if !exist {
		return nil, fmt.Errorf("Config is empty!")
	}

	config, err := getConfig()
	if err != nil {
		return nil, err
	} else if len(config.Caches) == 0 {
		return nil, fmt.Errorf("Config is empty!")
	}

	for _, v := range config.Caches {
		// Skipping the error, I know....
		if is, _ := IsSameUrl(url, v.Url); is {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("Url not found in cache: %v", url)
}

func getConfig() (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func InitConfig() (exist bool, err error) {
	os.Mkdir(configFolder, 0755)

	f, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return true, nil
		}
		return false, err
	}
	defer f.Close()

	c := DefaultConfig

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return false, err
	}

	_, err = f.Write(data)
	return false, err
}

// func (c *Config) Clear()
