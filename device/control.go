package device

import (
	"errors"
	"fmt"
	"time"
)

// InputCapable is implemented by devices that accept native input.
// Coordinates are in screenshot pixels, as returned by GetScreenshot.
type InputCapable interface {
	Tap(x, y int) error
	Swipe(fromX, fromY, toX, toY int, duration time.Duration) error
	InputText(text string) error
	// PressKey presses a named key ("home", "back", "enter", "menu", "delete",
	// "tab", "app_switch", "power", "volume_up", "volume_down") or, where the
	// platform supports it, a numeric key code.
	PressKey(key string) error
}

// LogReader is implemented by devices that can return their recent system log.
type LogReader interface {
	// ReadLogs returns up to lines of the most recent log lines, oldest first.
	ReadLogs(lines int) ([]string, error)
}

// Control operations, used to forward native control calls to node devices.
const (
	ControlTap      = "tap"
	ControlSwipe    = "swipe"
	ControlText     = "text"
	ControlKey      = "key"
	ControlReadLogs = "logs"
)

// ControlRequest is a native control call in a transport-friendly form.
type ControlRequest struct {
	DeviceID   string
	Op         string
	X, Y       int
	ToX, ToY   int
	DurationMs int
	Text       string
	Key        string
	Lines      int
}

type ControlResponse struct {
	Lines []string
}

// ApplyControl executes the request on a local device.
func ApplyControl(dev Device, req ControlRequest) (ControlResponse, error) {
	var resp ControlResponse
	switch req.Op {
	case ControlTap, ControlSwipe, ControlText, ControlKey:
		input, ok := dev.(InputCapable)
		if !ok {
			return resp, ErrUnsupported
		}
		switch req.Op {
		case ControlTap:
			return resp, input.Tap(req.X, req.Y)
		case ControlSwipe:
			return resp, input.Swipe(req.X, req.Y, req.ToX, req.ToY, time.Duration(req.DurationMs)*time.Millisecond)
		case ControlText:
			return resp, input.InputText(req.Text)
		default:
			return resp, input.PressKey(req.Key)
		}
	case ControlReadLogs:
		reader, ok := dev.(LogReader)
		if !ok {
			return resp, ErrUnsupported
		}
		lines, err := reader.ReadLogs(req.Lines)
		resp.Lines = lines
		return resp, err
	}
	return resp, fmt.Errorf("unknown control operation %q", req.Op)
}

// ControlError restores well-known errors from their message after a remote
// call, so callers can still use errors.Is.
func ControlError(message string) error {
	if message == ErrUnsupported.Error() {
		return ErrUnsupported
	}
	return errors.New(message)
}
