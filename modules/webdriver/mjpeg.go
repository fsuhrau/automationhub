package webdriver

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// DefaultMjpegServerPort is the port of the WebDriverAgent MJPEG server
// (MJPEG_SERVER_PORT on the runner).
const DefaultMjpegServerPort = 9100

// maxMjpegFrame bounds one JPEG frame of the MJPEG stream.
const maxMjpegFrame = 32 << 20

// ScreenInfo is the answer of /wda/screen.
type ScreenInfo struct {
	// Width and Height are in points, in the current orientation.
	Width  float64
	Height float64
	// Scale is the number of screenshot pixels per point.
	Scale float64
}

// Screen returns the screen size in points and the scale of the device.
func (c *Client) Screen() (ScreenInfo, error) {
	var result struct {
		Value struct {
			ScreenSize struct {
				Width  float64 `json:"width"`
				Height float64 `json:"height"`
			} `json:"screenSize"`
			Scale float64 `json:"scale"`
		} `json:"value"`
	}
	if err := c.call(call{"get screen", http.MethodGet, c.getUrl("/wda/screen"), nil}, &result); err != nil {
		return ScreenInfo{}, err
	}
	info := ScreenInfo{Width: result.Value.ScreenSize.Width, Height: result.Value.ScreenSize.Height, Scale: result.Value.Scale}
	if info.Scale <= 0 {
		info.Scale = 1
	}
	return info, nil
}

// UpdateSettings changes Appium settings of the session, e.g.
// mjpegServerFramerate. WebDriverAgent applies them to the whole runner.
func (c *Client) UpdateSettings(settings map[string]any) error {
	return c.post("update settings", "/appium/settings", map[string]any{"settings": settings})
}

// Alive reports whether the WebDriverAgent server answers.
func (c *Client) Alive(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.do(ctx, call{"get status", http.MethodGet, c.getUrl("/status"), nil}, nil) == nil
}

// MjpegURL is the URL of the MJPEG server on the host of the WebDriverAgent.
func (c *Client) MjpegURL(port int) (string, error) {
	u, err := url.Parse(c.address)
	if err != nil {
		return "", err
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("no host in WebDriverAgent address %q", c.address)
	}
	return (&url.URL{Scheme: "http", Host: net.JoinHostPort(u.Hostname(), strconv.Itoa(port)), Path: "/"}).String(), nil
}

// MjpegStream reads the JPEG frames of the WebDriverAgent MJPEG server.
type MjpegStream struct {
	body io.ReadCloser
	r    *MjpegReader
}

// OpenMjpeg connects to an MJPEG server. The stream ends when ctx is done or
// it is closed.
func OpenMjpeg(ctx context.Context, address string) (*MjpegStream, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	// no client timeout, the response is endless; only the connect is bounded
	client := http.Client{Transport: &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
		ResponseHeaderTimeout: 10 * time.Second,
		DisableCompression:    true,
	}}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connect to MJPEG server %s: %w", address, err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("MJPEG server %s: %s", address, resp.Status)
	}
	return &MjpegStream{body: resp.Body, r: NewMjpegReader(resp.Body)}, nil
}

// ReadFrame returns the next JPEG frame.
func (s *MjpegStream) ReadFrame() ([]byte, error) { return s.r.ReadFrame() }

func (s *MjpegStream) Close() error { return s.body.Close() }
