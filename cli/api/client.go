package api

import (
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/endpoints/api"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httputil"
	"time"
)

type Client struct {
	BaseURL    string
	apiToken   string
	HTTPClient *http.Client
}

func NewClient(url, apiToken, projectID string, appId uint) *Client {
	return &Client{
		BaseURL:  fmt.Sprintf("%s/api/%s/app/%d", url, projectID, appId),
		apiToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("X-Auth-Token", c.apiToken)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes api.ErrorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}

		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	httputil.DumpRequest(req, false)

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}

	return nil
}
