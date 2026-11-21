package androiddevice

import (
	"fmt"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/androiddevice/scrcpy"
)

// injector is the control channel of a scrcpy session.
type injector interface {
	InjectTouch(action int, pointerID int64, pos scrcpy.Position) error
	InjectKeycode(action, keycode int) error
	InjectText(text string) error
	InjectScroll(pos scrcpy.Position, hScroll, vScroll float64) error
	ResetVideo() error
}

// videoStream is the device.VideoStream and device.VideoInput of a session.
type videoStream struct {
	videoSession
	meta device.VideoMeta
}

func (s *videoStream) Meta() device.VideoMeta { return s.meta }

// InjectInput implements device.VideoInput through the scrcpy control socket.
func (s *videoStream) InjectInput(ev device.InputEvent) error {
	switch ev.Kind {
	case device.InputTouch:
		return s.injectTouch(ev)
	case device.InputKey:
		return s.injectKey(ev)
	case device.InputText:
		return s.InjectText(ev.Text)
	case device.InputScroll:
		return s.InjectScroll(position(ev), ev.ScrollX, ev.ScrollY)
	case device.InputResetVideo:
		return s.ResetVideo()
	}
	return fmt.Errorf("unknown input kind %q", ev.Kind)
}

func position(ev device.InputEvent) scrcpy.Position {
	return scrcpy.Position{X: ev.X, Y: ev.Y, Width: ev.Width, Height: ev.Height}
}

func (s *videoStream) injectTouch(ev device.InputEvent) error {
	action, ok := touchAction(ev.Action)
	if !ok {
		return fmt.Errorf("unknown touch action %q", ev.Action)
	}
	if ev.Width <= 0 || ev.Height <= 0 {
		return fmt.Errorf("touch needs the video width and height")
	}
	return s.InjectTouch(action, ev.PointerID, position(ev))
}

func touchAction(name string) (int, bool) {
	actions := map[string]int{
		device.InputDown:   scrcpy.ActionDown,
		device.InputMove:   scrcpy.ActionMove,
		device.InputUp:     scrcpy.ActionUp,
		device.InputCancel: scrcpy.ActionCancel,
	}
	action, ok := actions[name]
	return action, ok
}

func (s *videoStream) injectKey(ev device.InputEvent) error {
	code, err := keyCode(ev.Key)
	if err != nil {
		return err
	}
	switch ev.Action {
	case device.InputDown:
		return s.InjectKeycode(scrcpy.ActionDown, code)
	case device.InputUp:
		return s.InjectKeycode(scrcpy.ActionUp, code)
	case device.InputPress, "":
		return s.pressKey(code)
	}
	return fmt.Errorf("unknown key action %q", ev.Action)
}

// pressKey sends a key down followed by a key up.
func (s *videoStream) pressKey(code int) error {
	if err := s.InjectKeycode(scrcpy.ActionDown, code); err != nil {
		return err
	}
	return s.InjectKeycode(scrcpy.ActionUp, code)
}
