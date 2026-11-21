package webdriver

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

// Screenshot returns the PNG of the screen.
func (c *Client) Screenshot() ([]byte, error) {
	var result struct {
		Value string `json:"value"`
	}
	if err := c.call(call{"take screenshot", http.MethodGet, c.getUrl("/screenshot"), nil}, &result); err != nil {
		return nil, err
	}
	png, err := base64.StdEncoding.DecodeString(result.Value)
	if err != nil {
		return nil, fmt.Errorf("unable to take screenshot: invalid image: %w", err)
	}
	return png, nil
}
