package main

import (
	"net/http"
)

type HTTPClient struct {
	http.Client
}
type HTTPClientConfig struct {
	UserAgent string
}

// NewHTTPClient is a wrapper for http client. Here we configuring http client once to use it in the application.
func NewHTTPClient() (*HTTPClient, *HTTPClientConfig) {
	client := &HTTPClient{}
	return client, &HTTPClientConfig{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
	}
}

// NewReq is a wrapper for http NewRequest method to set headers once
func (c *HTTPClient) NewHTTPRequest(method, endpoint string, body interface{}) *http.Request {
	req, _ := http.NewRequest(method, endpoint, nil)
	return req
}
