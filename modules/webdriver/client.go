package webdriver

import (
	"net/http"
)

type Client struct {
	httpClient *http.Client
	address    string
	sessionId  string
}

func New(address string) *Client {
	return &Client{
		httpClient: &http.Client{},
		address:    address,
	}
}
