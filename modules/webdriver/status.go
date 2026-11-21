package webdriver

import "net/http"

type Build struct {
	Time                    string `json:"time"`
	ProductBundleIdentifier string `json:"productBundleIdentifier"`
}

type StatusValue struct {
	Message string            `json:"message"`
	State   string            `json:"state"`
	OS      map[string]string `json:"os"`
	IOS     map[string]string `json:"ios"`
	Ready   bool              `json:"ready"`
	Build   Build             `json:"build"`
}

type Status struct {
	SessionID string      `json:"sessionId"`
	Value     StatusValue `json:"value"`
}

// Status asks the server for its state.
func (c *Client) Status() (*Status, error) {
	var status Status
	if err := c.call(call{"get status", http.MethodGet, c.getUrl("/status"), nil}, &status); err != nil {
		return nil, err
	}
	return &status, nil
}
