package scrcpy

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"
)

// Control message types of the scrcpy protocol (ControlMessage.java).
const (
	typeInjectKeycode = 0
	typeInjectText    = 1
	typeInjectTouch   = 2
	typeInjectScroll  = 3
	typeResetVideo    = 17
)

// Android MotionEvent and KeyEvent actions.
const (
	ActionDown   = 0
	ActionUp     = 1
	ActionMove   = 2
	ActionCancel = 3
)

// injectTextMaxLength is the server's limit for one text message in bytes.
const injectTextMaxLength = 300

const controlWriteTimeout = 2 * time.Second

// ErrNoControl is returned when the session was started without control.
var ErrNoControl = errors.New("scrcpy session has no control channel")

// pointerIDMouse makes the server inject a mouse event; any other id is a finger.
const pointerIDMouse = -1

func encodeTouch(action int, pointerID int64, x, y, width, height int) []byte {
	b := make([]byte, 32)
	b[0] = typeInjectTouch
	b[1] = byte(action)
	binary.BigEndian.PutUint64(b[2:], uint64(pointerID))
	putPosition(b[10:], x, y, width, height)
	if action != ActionUp {
		binary.BigEndian.PutUint16(b[22:], 0xffff) // pressure 1.0
	}
	// action button and buttons stay 0: buttons must not be set for touches
	return b
}

func encodeKeycode(action, keycode int) []byte {
	b := make([]byte, 14)
	b[0] = typeInjectKeycode
	b[1] = byte(action)
	binary.BigEndian.PutUint32(b[2:], uint32(keycode))
	// repeat and meta state stay 0
	return b
}

func encodeText(text string) ([]byte, error) {
	if len(text) > injectTextMaxLength {
		return nil, fmt.Errorf("text is longer than %d bytes", injectTextMaxLength)
	}
	b := make([]byte, 5+len(text))
	b[0] = typeInjectText
	binary.BigEndian.PutUint32(b[1:], uint32(len(text)))
	copy(b[5:], text)
	return b, nil
}

// encodeScroll encodes a scroll; hScroll and vScroll are in [-16, 16] "lines".
func encodeScroll(x, y, width, height int, hScroll, vScroll float64) []byte {
	b := make([]byte, 21)
	b[0] = typeInjectScroll
	putPosition(b[1:], x, y, width, height)
	binary.BigEndian.PutUint16(b[13:], uint16(i16FixedPoint(hScroll/16)))
	binary.BigEndian.PutUint16(b[15:], uint16(i16FixedPoint(vScroll/16)))
	return b
}

func putPosition(b []byte, x, y, width, height int) {
	binary.BigEndian.PutUint32(b, uint32(int32(x)))
	binary.BigEndian.PutUint32(b[4:], uint32(int32(y)))
	binary.BigEndian.PutUint16(b[8:], uint16(width))
	binary.BigEndian.PutUint16(b[10:], uint16(height))
}

// i16FixedPoint converts [-1, 1] to a 16 bit fixed point value (Binary.java).
func i16FixedPoint(f float64) int16 {
	f = math.Max(-1, math.Min(1, f))
	if f == 1 {
		return math.MaxInt16
	}
	return int16(f * 0x8000)
}

// InjectTouch injects a finger event. x and y are pixels of a video frame of
// width x height, which must be the current video size; the server drops
// events for another size (e.g. while rotating).
func (s *Session) InjectTouch(action int, pointerID int64, x, y, width, height int) error {
	if pointerID == pointerIDMouse {
		pointerID = -2 // the generic finger
	}
	return s.writeControl(encodeTouch(action, pointerID, x, y, width, height))
}

// InjectKeycode injects an Android key event.
func (s *Session) InjectKeycode(action, keycode int) error {
	return s.writeControl(encodeKeycode(action, keycode))
}

// InjectText types text (at most 300 bytes per call).
func (s *Session) InjectText(text string) error {
	b, err := encodeText(text)
	if err != nil {
		return err
	}
	return s.writeControl(b)
}

// InjectScroll injects a scroll at a position, like a mouse wheel.
func (s *Session) InjectScroll(x, y, width, height int, hScroll, vScroll float64) error {
	return s.writeControl(encodeScroll(x, y, width, height, hScroll, vScroll))
}

// ResetVideo asks the server to restart the encoder, which sends a new
// config packet and key frame (e.g. after a decoder error).
func (s *Session) ResetVideo() error {
	return s.writeControl([]byte{typeResetVideo})
}

func (s *Session) writeControl(b []byte) error {
	if s.control == nil {
		return ErrNoControl
	}
	s.controlMu.Lock()
	defer s.controlMu.Unlock()
	_ = s.control.SetWriteDeadline(time.Now().Add(controlWriteTimeout))
	_, err := s.control.Write(b)
	return err
}
