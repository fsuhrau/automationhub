package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fsuhrau/automationhub/endpoints/httpx"
)

const jsonContentType = "application/json; charset=utf-8"

type Client struct {
	BaseURL    string
	apiToken   string
	HTTPClient *http.Client
}

// Config says which app of which hub the client talks to.
type Config struct {
	URL       string
	Token     string
	ProjectID string
	AppID     uint
}

func NewClient(cfg Config) *Client {
	return &Client{
		BaseURL:  fmt.Sprintf("%s/api/%s/app/%d", cfg.URL, cfg.ProjectID, cfg.AppID),
		apiToken: cfg.Token,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

// jsonCall is a request with a JSON body (when In is set) and a JSON answer.
type jsonCall struct {
	method string
	path   string
	in     any
	out    any
}

// rawCall is a request with a prepared body.
type rawCall struct {
	method      string
	path        string
	contentType string
	body        io.Reader
}

// sendJSON sends the call's body as JSON and decodes the answer into its out.
func (c *Client) sendJSON(ctx context.Context, call jsonCall) error {
	var body io.Reader
	if call.in != nil {
		data, err := json.Marshal(call.in)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(data)
	}
	return c.send(ctx, rawCall{call.method, call.path, jsonContentType, body}, call.out)
}

// send does one authenticated request and decodes the JSON answer into out.
func (c *Client) send(ctx context.Context, call rawCall, out any) error {
	req, err := http.NewRequestWithContext(ctx, call.method, c.BaseURL+call.path, call.body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", call.contentType)
	req.Header.Set("Accept", jsonContentType)
	req.Header.Set("X-Auth-Token", c.apiToken)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return responseError(res)
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// responseError is the message of the hub's error answer, or the status code.
func responseError(res *http.Response) error {
	var errRes httpx.ErrorResponse
	if err := json.NewDecoder(res.Body).Decode(&errRes); err == nil {
		return errors.New(errRes.Message)
	}
	return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
}
