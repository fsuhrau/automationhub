package webdriver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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

func (c *Client) Status() (*Status, error) {
	req, err := http.NewRequest("GET", c.getUrl("/status"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(resp.Body)

	var status Status
	if err := json.Unmarshal(body, &status); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(body, &wdaError)
		return nil, fmt.Errorf("unable to get status: %s", wdaError.Value.Error)
	}
	return &status, err
}
