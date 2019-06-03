package tunnel_client

import (
	"crypto/tls"
	"net/http"
	"time"
)

var HttpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost:   1024,
		IdleConnTimeout:       0,
		ExpectContinueTimeout: 1 * time.Second,
		TLSHandshakeTimeout:   0,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Proxy: nil,
	},
}
