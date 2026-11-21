package iosdevice

import (
	"context"
	"time"

	"github.com/danielpaulus/go-ios/ios"
)

// muxListener is an open usbmuxd "Listen" subscription.
type muxListener interface {
	// Next blocks until a device was attached or detached.
	Next() (attached bool, serial string, err error)
	Close()
}

// openUsbmuxListener subscribes to usbmuxd device events.
func openUsbmuxListener() (muxListener, error) {
	conn, err := ios.NewUsbMuxConnectionSimple()
	if err != nil {
		return nil, err
	}
	next, err := conn.Listen()
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &usbmuxListener{conn: conn, next: next}, nil
}

type usbmuxListener struct {
	conn *ios.UsbMuxConnection
	next func() (ios.AttachedMessage, error)
}

func (l *usbmuxListener) Next() (bool, string, error) {
	msg, err := l.next()
	if err != nil {
		return false, "", err
	}
	return msg.DeviceAttached(), msg.Properties.SerialNumber, nil
}

func (l *usbmuxListener) Close() { l.conn.Close() }

// usbEvents are the collaborators of watchUsbmux.
type usbEvents struct {
	open func() (muxListener, error)
	// onChange is called for every event and once after each connect.
	onChange func()
	// connected reports whether the subscription is up.
	connected func(bool)
}

// watchUsbmux calls onChange for every attach/detach event and once after
// each (re)connect, so changes during a disconnect aren't missed. It
// reconnects with backoff until ctx is done.
func watchUsbmux(ctx context.Context, ev usbEvents) {
	backoff := time.Second
	for ctx.Err() == nil {
		if listener, err := ev.open(); err == nil {
			backoff = time.Second
			ev.serve(ctx, listener)
		}
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

// serve reports the events of an open subscription until it breaks or ctx is
// done.
func (ev usbEvents) serve(ctx context.Context, listener muxListener) {
	ev.connected(true)
	ev.onChange()
	stop := context.AfterFunc(ctx, listener.Close)
	for {
		if _, _, err := listener.Next(); err != nil {
			break
		}
		ev.onChange()
	}
	stop()
	listener.Close()
	ev.connected(false)
}
