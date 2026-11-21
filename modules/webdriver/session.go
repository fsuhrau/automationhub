package webdriver

import "net/http"

type SessionValue struct {
	SessionID    string            `json:"sessionId"`
	Capabilities map[string]string `json:"capabilities"`
}

type Session struct {
	SessionID string       `json:"sessionId"`
	Value     SessionValue `json:"value"`
}

// CreateSession opens a session; later calls of the client belong to it.
func (c *Client) CreateSession() (*Session, error) {
	// we don't need capabilities for now
	body := map[string]any{"capabilities": map[string]any{}}
	var session Session
	if err := c.call(call{"create session", http.MethodPost, c.getUrl("/session"), body}, &session); err != nil {
		return nil, err
	}
	c.sessionId = session.SessionID
	return &session, nil
}

// CloseSession ends the session of the client.
func (c *Client) CloseSession() error {
	err := c.call(call{"close session", http.MethodDelete, c.getUrl("/session"), nil}, nil)
	c.sessionId = ""
	return err
}
