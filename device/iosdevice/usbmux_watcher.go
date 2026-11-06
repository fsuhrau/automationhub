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

// watchUsbmux calls onChange for every attach/detach event and once after
// each (re)connect, so changes during a disconnect aren't missed. It
// reconnects with backoff until ctx is done. connected reports whether the
// subscription is up.
func watchUsbmux(ctx context.Context, open func() (muxListener, error), onChange func(), connected func(bool)) {
	backoff := time.Second
	for ctx.Err() == nil {
		listener, err := open()
		if err == nil {
			backoff = time.Second
			connected(true)
			onChange()
			stop := context.AfterFunc(ctx, listener.Close)
			for {
				if _, _, err = listener.Next(); err != nil {
					break
				}
				onChange()
			}
			stop()
			listener.Close()
			connected(false)
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
