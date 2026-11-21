package webdriver

import (
	"errors"
	"net/http"
)

type alertResponse struct {
	SessionId string `json:"sessionId"`
	Value     string `json:"value"`
}

// GetAlertText returns the text of the visible alert.
func (c *Client) GetAlertText() (string, error) {
	var r alertResponse
	if err := c.call(call{"get alert text", http.MethodGet, c.getSessionUrl("/alert/text"), nil}, &r); err != nil {
		return "", err
	}
	if len(r.Value) == 0 {
		return "", errors.New("there is no alert text")
	}
	return r.Value, nil
}

// AcceptAlert accepts the visible alert, or dismisses it when accept is false.
func (c *Client) AcceptAlert(accept bool) (string, error) {
	action := "accept"
	if !accept {
		action = "dismiss"
	}
	var r alertResponse
	if err := c.call(call{"accept/dismiss alert", http.MethodPost, c.getSessionUrl("/alert/%s", action), nil}, &r); err != nil {
		return "", err
	}
	return r.Value, nil
}
