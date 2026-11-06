package control

import (
	"context"
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/manager"
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

func inputOf(dev device.Device) (device.InputCapable, error) {
	input, ok := dev.(device.InputCapable)
	if !ok {
		return nil, device.ErrUnsupported
	}
	return input, nil
}

// --- read-only, no lease needed ---

// Screenshot returns a native PNG screenshot.
func (s *Service) Screenshot(dev device.Device) ([]byte, int, int, error) {
	return dev.GetScreenshot()
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

// --- control, lease required ---

func (s *Service) Tap(dev device.Device, token string, x, y int) error {
	if err := s.Authorize(dev.DeviceID(), token); err != nil {
		return err
	}
	input, err := inputOf(dev)
	if err != nil {
		return err
	}
	return input.Tap(x, y)
}

func (s *Service) Swipe(dev device.Device, token string, fromX, fromY, toX, toY int, duration time.Duration) error {
	if err := s.Authorize(dev.DeviceID(), token); err != nil {
		return err
	}
	input, err := inputOf(dev)
	if err != nil {
		return err
	}
	return input.Swipe(fromX, fromY, toX, toY, duration)
}

func (s *Service) InputText(dev device.Device, token, text string) error {
	if err := s.Authorize(dev.DeviceID(), token); err != nil {
		return err
	}
	input, err := inputOf(dev)
	if err != nil {
		return err
	}
	return input.InputText(text)
}

func (s *Service) PressKey(dev device.Device, token, key string) error {
	if err := s.Authorize(dev.DeviceID(), token); err != nil {
		return err
	}
	input, err := inputOf(dev)
	if err != nil {
		return err
	}
	return input.PressKey(key)
}

// TouchElement touches an app element through the Unity plugin.
func (s *Service) TouchElement(ctx context.Context, dev device.Device, token string, ref ElementRef) error {
	if err := s.Authorize(dev.DeviceID(), token); err != nil {
		return err
	}
	id, err := s.resolveElement(ctx, dev, ref)
	if err != nil {
		return err
	}
	a := action.TouchElement{ElementID: id}
	if err := s.execute(ctx, dev, &a); err != nil {
		return err
	}
	if !a.Success {
		return fmt.Errorf("app could not touch element %s", id)
	}
	return nil
}

// SetElementValue sets an attribute (e.g. "text") of an app element.
func (s *Service) SetElementValue(ctx context.Context, dev device.Device, token string, ref ElementRef, attr, value string) error {
	if err := s.Authorize(dev.DeviceID(), token); err != nil {
		return err
	}
	id, err := s.resolveElement(ctx, dev, ref)
	if err != nil {
		return err
	}
	a := action.SetValue{ElementID: id, Attr: attr, Value: value}
	if err := s.execute(ctx, dev, &a); err != nil {
		return err
	}
	if !a.Success {
		return fmt.Errorf("app could not set %s on element %s", attr, id)
	}
	return nil
}

// InstallApp installs the app; for node devices the bundle is uploaded to the
// node first.
func (s *Service) InstallApp(ctx context.Context, dev device.Device, token string, params *app.Parameter) error {
	if err := s.Authorize(dev.DeviceID(), token); err != nil {
		return err
	}
	if nd, ok := dev.(nodeDevice); ok {
		if err := nd.NodeManager().UploadApp(ctx, nd.GetNodeID(), params); err != nil {
			return fmt.Errorf("upload app to node: %w", err)
		}
	}
	return dev.InstallApp(params)
}

// StartApp starts the app. The app connects back to nodeURL with sessionID.
func (s *Service) StartApp(dev device.Device, token string, params *app.Parameter, sessionID, nodeURL string) error {
	if err := s.Authorize(dev.DeviceID(), token); err != nil {
		return err
	}
	return dev.StartApp(nil, params, sessionID, nodeURL)
}

// StopApp stops the app and, when clearData is set and supported, wipes its data.
func (s *Service) StopApp(dev device.Device, token string, params *app.Parameter, clearData bool) error {
	if err := s.Authorize(dev.DeviceID(), token); err != nil {
		return err
	}
	if err := dev.StopApp(params); err != nil {
		return err
	}
	if clearData {
		clearer, ok := dev.(device.AppDataClearer)
		if !ok {
			return fmt.Errorf("clear app data: %w", device.ErrUnsupported)
		}
		return clearer.ClearAppData(params)
	}
	return nil
}

// --- helpers ---

type nodeDevice interface {
	GetNodeID() manager.NodeIdentifier
	NodeManager() manager.Nodes
}

func (s *Service) execute(ctx context.Context, dev device.Device, a action.Interface) error {
	if !dev.IsAppConnected() {
		return fmt.Errorf("%w (%s); start the app first", ErrAppNotConnected, dev.DeviceID())
	}
	return tester_action.NewExecutor(s.dm).Execute(ctx, dev, a, s.actionTimeout)
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
