package goreqx

import (
	Url "net/url"
	"strings"
)

func buildUrl(url string) (string, error) {
	u, err := Url.Parse(url)
	if err != nil {
		return "", err
	}

	if u.Path == "" {
		u.Path = "/"
	}

	return u.Scheme + "://" + u.Host + u.Path, nil
}

func IsSameUrl(url1 string, url2 string) (bool, error) {
	u1, err := buildUrl(url1)
	if err != nil {
		return false, err
	}

	u2, err := buildUrl(url2)
	if err != nil {
		return false, err
	}

	return strings.EqualFold(u1, u2), nil
}