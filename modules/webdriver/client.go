package webdriver

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
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
