package webdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/session", c.address), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	responseBody, _ := ioutil.ReadAll(resp.Body)

	var session Session
	err = json.Unmarshal(responseBody, &session)
	c.sessionId = session.SessionID
	return &session, err
}

func (c *Client) CloseSession() error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/session", c.address), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	_, err = c.httpClient.Do(req)
	return err
}
