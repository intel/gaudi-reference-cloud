package utils

import (
	"net/url"
)

func ConstructURL(base_url string, params map[string]string) (string, error) {
	constructed_url, err := url.Parse(base_url)
	if err != nil {
		return "", err
	}

	if params != nil && len(params) > 0 {
		q := constructed_url.Query()
		for key, value := range params {
			q.Add(key, value)
		}
		constructed_url.RawQuery = q.Encode()
	}

	return constructed_url.String(), nil
}
