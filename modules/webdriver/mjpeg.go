package webdriver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	resp, err := c.httpClient.Get(c.getUrl("/wda/screen"))
	if err != nil {
		return ScreenInfo{}, err
	}
	defer resp.Body.Close()
	var result struct {
		Value struct {
			ScreenSize struct {
				Width  float64 `json:"width"`
				Height float64 `json:"height"`
			} `json:"screenSize"`
			Scale float64 `json:"scale"`
		} `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ScreenInfo{}, fmt.Errorf("/wda/screen: %w", err)
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
	_, _, err := c.postJSON("/appium/settings", map[string]any{"settings": settings})
	return err
}

// Alive reports whether the WebDriverAgent server answers.
func (c *Client) Alive(timeout time.Duration) bool {
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(c.getUrl("/status"))
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
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

// MjpegReader splits a multipart/x-mixed-replace body into JPEG frames.
//
// It doesn't use mime/multipart: WebDriverAgent declares the boundary as
// "--BoundaryString" but writes "--BoundaryString" as the delimiter, which a
// strict parser rejects. Parts are found by their Content-Length header;
// parts without one are read up to the JPEG end marker.
type MjpegReader struct {
	r *bufio.Reader
}

func NewMjpegReader(r io.Reader) *MjpegReader {
	return &MjpegReader{r: bufio.NewReaderSize(r, 64<<10)}
}

// ReadFrame returns the next frame. The slice belongs to the caller.
func (m *MjpegReader) ReadFrame() ([]byte, error) {
	for {
		length, err := m.readPartHeader()
		if err != nil {
			return nil, err
		}
		var frame []byte
		if length >= 0 {
			if length > maxMjpegFrame {
				return nil, fmt.Errorf("MJPEG frame of %d bytes is too large", length)
			}
			frame = make([]byte, length)
			if _, err := io.ReadFull(m.r, frame); err != nil {
				return nil, err
			}
		} else if frame, err = m.readJPEG(); err != nil {
			return nil, err
		}
		if len(frame) > 0 {
			return frame, nil
		}
	}
}

// readPartHeader skips to the next part and reads its headers. It returns the
// Content-Length, or -1 when the part has none.
func (m *MjpegReader) readPartHeader() (int, error) {
	// skip blank lines and the boundary up to the first header line
	var line string
	for {
		l, err := m.readLine()
		if err != nil {
			return 0, err
		}
		if l == "" || strings.HasPrefix(l, "--") {
			continue
		}
		line = l
		break
	}
	length := -1
	for line != "" {
		name, value, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			n, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || n < 0 {
				return 0, fmt.Errorf("invalid MJPEG Content-Length %q", value)
			}
			length = n
		}
		l, err := m.readLine()
		if err != nil {
			return 0, err
		}
		line = l
	}
	return length, nil
}

func (m *MjpegReader) readLine() (string, error) {
	line, err := m.r.ReadSlice('\n')
	if err == bufio.ErrBufferFull {
		return "", fmt.Errorf("MJPEG header line too long")
	}
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(line), "\r\n"), nil
}

// readJPEG reads from the start of image (FF D8) to the end of image (FF D9)
// marker.
func (m *MjpegReader) readJPEG() ([]byte, error) {
	var buf bytes.Buffer
	var prev byte
	started := false
	for {
		b, err := m.r.ReadByte()
		if err != nil {
			return nil, err
		}
		if !started {
			if prev == 0xFF && b == 0xD8 {
				started = true
				buf.Write([]byte{0xFF, 0xD8})
			}
			prev = b
			continue
		}
		buf.WriteByte(b)
		if buf.Len() > maxMjpegFrame {
			return nil, fmt.Errorf("MJPEG frame is too large")
		}
		if prev == 0xFF && b == 0xD9 {
			return buf.Bytes(), nil
		}
		prev = b
	}
}
