package webdriver

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
)

// Client talks to a WebDriverAgent server.
type Client struct {
	httpClient *http.Client
	address    string
	sessionId  string
}

// New creates a client of the WebDriverAgent at address. It sends its requests
// with httpClient; a nil httpClient gets a client without timeout.
func New(address string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{httpClient: httpClient, address: address}
}

func (c *Client) getUrl(format string, data ...any) string {
	u, _ := url.Parse(c.address)
	u.Path = path.Join(u.Path, fmt.Sprintf(format, data...))
	return u.String()
}

func (c *Client) getSessionUrl(format string, data ...any) string {
	u, _ := url.Parse(c.address)
	u.Path = path.Join(u.Path, "session", c.sessionId, fmt.Sprintf(format, data...))
	return u.String()
}
