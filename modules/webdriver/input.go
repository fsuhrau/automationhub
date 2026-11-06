package webdriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// postJSON posts body to the session scoped path and returns the HTTP status
// and response body. A non-2xx status is returned as an error that carries the
// WDA error message when there is one.
func (c *Client) postJSON(path string, body any) (int, []byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return 0, nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.getSessionUrl("%s", path), bytes.NewReader(payload))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var wdaError WDAError
		_ = json.Unmarshal(respBody, &wdaError)
		msg := wdaError.Value.Message
		if msg == "" {
			msg = wdaError.Value.Error
		}
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return resp.StatusCode, respBody, fmt.Errorf("%s: %s", path, msg)
	}
	return resp.StatusCode, respBody, nil
}

// Tap taps at the given point (in WDA points, not pixels).
func (c *Client) Tap(x, y float64) error {
	body := map[string]float64{"x": x, "y": y}
	status, _, err := c.postJSON("/wda/tap", body)
	if err == nil {
		return nil
	}
	if status == http.StatusNotFound {
		// older WebDriverAgent builds only know /wda/tap/:uuid
		_, _, err = c.postJSON("/wda/tap/0", body)
	}
	return err
}

// Drag drags from one point to another (in WDA points) over duration seconds.
func (c *Client) Drag(fromX, fromY, toX, toY, durationSeconds float64) error {
	_, _, err := c.postJSON("/wda/dragfromtoforduration", map[string]float64{
		"fromX":    fromX,
		"fromY":    fromY,
		"toX":      toX,
		"toY":      toY,
		"duration": durationSeconds,
	})
	return err
}

// ScreenScale returns the ratio between screenshot pixels and WDA points.
func (c *Client) ScreenScale() (float64, error) {
	resp, err := c.httpClient.Get(c.getUrl("/wda/screen"))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var result struct {
		Value struct {
			Scale float64 `json:"scale"`
		} `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	if result.Value.Scale <= 0 {
		return 1, nil
	}
	return result.Value.Scale, nil
}
