package control

import (
	"context"
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
)

// --- control, lease required ---

// SwipeGesture is a swipe from one point to another.
type SwipeGesture struct {
	FromX, FromY, ToX, ToY int
	Duration               time.Duration
}

// ElementChange sets an attribute (e.g. "text") of an app element.
type ElementChange struct {
	Ref   ElementRef
	Attr  string
	Value string
}

// InjectInput injects low latency input through a running video stream of
// the device. It needs the lease like the other control calls, except for a
// video reset, which only asks for a new key frame.
func (s *Service) InjectInput(t Target, stream device.VideoStream, ev device.InputEvent) error {
	if ev.Kind != device.InputResetVideo {
		if err := s.authorize(t); err != nil {
			return err
		}
	}
	input, ok := stream.(device.VideoInput)
	if !ok {
		return device.ErrUnsupported
	}
	return input.InjectInput(ev)
}

// Tap taps the screen.
func (s *Service) Tap(t Target, x, y int) error {
	input, err := s.input(t)
	if err != nil {
		return err
	}
	return input.Tap(x, y)
}

// Swipe swipes over the screen.
func (s *Service) Swipe(t Target, g SwipeGesture) error {
	input, err := s.input(t)
	if err != nil {
		return err
	}
	return input.Swipe(g.FromX, g.FromY, g.ToX, g.ToY, g.Duration)
}

// InputText types the text.
func (s *Service) InputText(t Target, text string) error {
	input, err := s.input(t)
	if err != nil {
		return err
	}
	return input.InputText(text)
}

// PressKey presses a key.
func (s *Service) PressKey(t Target, key string) error {
	input, err := s.input(t)
	if err != nil {
		return err
	}
	return input.PressKey(key)
}

// TouchElement touches an app element through the Unity plugin.
func (s *Service) TouchElement(ctx context.Context, t Target, ref ElementRef) error {
	if err := s.authorize(t); err != nil {
		return err
	}
	id, err := s.resolveElement(ctx, t.Device, ref)
	if err != nil {
		return err
	}
	a := action.TouchElement{ElementID: id}
	if err := s.execute(ctx, t.Device, &a); err != nil {
		return err
	}
	if !a.Success {
		return fmt.Errorf("app could not touch element %s", id)
	}
	return nil
}

// SetElementValue sets an attribute (e.g. "text") of an app element.
func (s *Service) SetElementValue(ctx context.Context, t Target, change ElementChange) error {
	if err := s.authorize(t); err != nil {
		return err
	}
	id, err := s.resolveElement(ctx, t.Device, change.Ref)
	if err != nil {
		return err
	}
	a := action.SetValue{ElementID: id, Attr: change.Attr, Value: change.Value}
	if err := s.execute(ctx, t.Device, &a); err != nil {
		return err
	}
	if !a.Success {
		return fmt.Errorf("app could not set %s on element %s", change.Attr, id)
	}
	return nil
}
