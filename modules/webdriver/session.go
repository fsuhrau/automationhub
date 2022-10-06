package webdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SessionValue struct {
	SessionID    string            `json:"sessionId"`
	Capabilities map[string]string `json:"capabilities"`
}

type Session struct {
	SessionID string       `json:"sessionId"`
	Value     SessionValue `json:"value"`
}

func (c *Client) CreateSession() (*Session, error) {

	// we don't need capabilities for now
	jsonRequest := []byte(`{"capabilities": {}}`)

	body := bytes.NewBuffer(jsonRequest)
	req, err := http.NewRequest("POST", c.getUrl("/session"), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	responseBody, _ := io.ReadAll(resp.Body)

	var session Session
	if err := json.Unmarshal(responseBody, &session); err != nil {
		var wdaError WDAError
		_ = json.Unmarshal(responseBody, &wdaError)
		return nil, fmt.Errorf("unable to create session: %s", wdaError.Value.Error)
	}
	c.sessionId = session.SessionID
	return &session, err
}

func (c *Client) CloseSession() error {
	req, err := http.NewRequest("DELETE", c.getUrl("/session"), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	_, err = c.httpClient.Do(req)
	c.sessionId = ""
	return err
}
