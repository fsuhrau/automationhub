package webdriver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// call describes one request to WebDriverAgent. action names it in errors
// ("create session" gives "unable to create session: …").
type call struct {
	action string
	method string
	url    string
	body   any // marshalled to JSON when not nil
}

// HTTPError is the error response of WebDriverAgent.
type HTTPError struct {
	Action  string
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("unable to %s: %s", e.Action, e.Message)
}

// call sends the request without deadline, see do.
func (c *Client) call(cl call, out any) error {
	return c.do(context.Background(), cl, out)
}

// do sends the request and decodes a successful JSON response into out, which
// may be nil. A response with a non-2xx status is an *HTTPError.
func (c *Client) do(ctx context.Context, cl call, out any) error {
	payload, err := encodeBody(cl.body)
	if err != nil {
		return fmt.Errorf("unable to %s: %w", cl.action, err)
	}
	req, err := http.NewRequestWithContext(ctx, cl.method, cl.url, payload)
	if err != nil {
		return fmt.Errorf("unable to %s: %w", cl.action, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to %s: %w", cl.action, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to %s: %w", cl.action, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &HTTPError{Action: cl.action, Status: resp.StatusCode, Message: errorMessage(body, resp.StatusCode)}
	}
	return decodeBody(cl.action, body, out)
}

func encodeBody(body any) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(payload), nil
}

func decodeBody(action string, body []byte, out any) error {
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("unable to %s: invalid response: %w", action, err)
	}
	return nil
}

// errorMessage picks the message of a WDAError, falling back to the HTTP status.
func errorMessage(body []byte, status int) string {
	var wdaError WDAError
	_ = json.Unmarshal(body, &wdaError)
	for _, msg := range []string{wdaError.Value.Message, wdaError.Value.Error} {
		if msg != "" {
			return msg
		}
	}
	return http.StatusText(status)
}
