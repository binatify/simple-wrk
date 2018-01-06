package loader

import (
	"net/http"
	"time"
)

func newClient() *http.Client {
	client := &http.Client{}

	client.Transport = &http.Transport{
		DisableCompression:    false,
		DisableKeepAlives:     false,
		ResponseHeaderTimeout: time.Second * time.Duration(5),
	}

	return client
}
