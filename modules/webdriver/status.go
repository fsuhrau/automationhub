package webdriver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/status", c.address), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var status Status
	err = json.Unmarshal(body, &status)
	return &status, err
}
