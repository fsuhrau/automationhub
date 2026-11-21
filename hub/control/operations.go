package control

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	tester_action "github.com/fsuhrau/automationhub/tester/action"
)

// ElementRef selects an app element either by its ID or by an XPath
// expression evaluated against the scene graph.
type ElementRef struct {
	ID    string `json:"elementId,omitempty"`
	XPath string `json:"xpath,omitempty"`
}

func (r ElementRef) IsZero() bool {
	return r.ID == "" && r.XPath == ""
}

// --- read-only, no lease needed ---

// Screenshot returns a native PNG screenshot.
func (s *Service) Screenshot(dev device.Device) ([]byte, int, int, error) {
	shooter, ok := device.As[device.Screenshotter](dev)
	if !ok {
		return nil, 0, 0, fmt.Errorf("screenshot: %w", device.ErrUnsupported)
	}
	return shooter.GetScreenshot()
}

// Video starts a live video stream of the screen. It ends when ctx is done
// or the stream is closed.
func (s *Service) Video(ctx context.Context, dev device.Device, opts device.VideoOptions) (device.VideoStream, error) {
	streamer, ok := dev.(device.VideoStreamer)
	if !ok {
		return nil, device.ErrUnsupported
	}
	return streamer.StartVideo(ctx, opts)
}

// AppScreenshot asks the connected app (Unity plugin) for a screenshot.
func (s *Service) AppScreenshot(ctx context.Context, dev device.Device) ([]byte, int, int, error) {
	var a action.GetScreenshot
	if err := s.execute(ctx, dev, &a); err != nil {
		return nil, 0, 0, err
	}
	data := a.ScreenshotData()
	if len(data) == 0 {
		return nil, 0, 0, fmt.Errorf("app returned no screenshot")
	}
	return data, a.Width(), a.Height(), nil
}

// Hierarchy returns the app's scene graph as XML.
func (s *Service) Hierarchy(ctx context.Context, dev device.Device, pretty bool) ([]byte, error) {
	var a action.GetSceenGraph
	if err := s.execute(ctx, dev, &a); err != nil {
		return nil, err
	}
	return a.XMLBytes(pretty)
}

// Logs returns the most recent device log lines.
func (s *Service) Logs(dev device.Device, lines int) ([]string, error) {
	reader, ok := dev.(device.LogReader)
	if !ok {
		return nil, device.ErrUnsupported
	}
	return reader.ReadLogs(lines)
}

// --- helpers ---

func (s *Service) execute(ctx context.Context, dev device.Device, a action.Interface) error {
	if !dev.IsAppConnected() {
		return fmt.Errorf("%w (%s); start the app first", ErrAppNotConnected, dev.DeviceID())
	}
	return tester_action.NewExecutor(s.dm).Execute(ctx, tester_action.Call{Device: dev, Action: a, Timeout: s.actionTimeout})
}

func (s *Service) resolveElement(ctx context.Context, dev device.Device, ref ElementRef) (string, error) {
	if ref.ID != "" {
		return ref.ID, nil
	}
	if ref.XPath == "" {
		return "", fmt.Errorf("element id or xpath is required")
	}
	var graph action.GetSceenGraph
	if err := s.execute(ctx, dev, &graph); err != nil {
		return "", err
	}
	return graph.FindElementID(ref.XPath)
}
