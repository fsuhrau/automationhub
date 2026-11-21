package webdriver

import (
	"errors"
	"net/http"
)

// post sends body to the session scoped path.
func (c *Client) post(action, path string, body any) error {
	return c.call(call{action, http.MethodPost, c.getSessionUrl("%s", path), body}, nil)
}

// Tap taps at the given point (in WDA points, not pixels).
func (c *Client) Tap(x, y float64) error {
	body := map[string]float64{"x": x, "y": y}
	err := c.post("tap", "/wda/tap", body)
	var httpErr *HTTPError
	if errors.As(err, &httpErr) && httpErr.Status == http.StatusNotFound {
		// older WebDriverAgent builds only know /wda/tap/:uuid
		return c.post("tap", "/wda/tap/0", body)
	}
	return err
}

// Point is a position on the screen in WDA points.
type Point struct{ X, Y float64 }

// Drag drags from one point to another over duration seconds.
func (c *Client) Drag(from, to Point, durationSeconds float64) error {
	return c.post("drag", "/wda/dragfromtoforduration", map[string]float64{
		"fromX":    from.X,
		"fromY":    from.Y,
		"toX":      to.X,
		"toY":      to.Y,
		"duration": durationSeconds,
	})
}

// ScreenScale returns the ratio between screenshot pixels and WDA points.
func (c *Client) ScreenScale() (float64, error) {
	screen, err := c.Screen()
	if err != nil {
		return 0, err
	}
	return screen.Scale, nil
}
