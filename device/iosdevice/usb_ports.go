package iosdevice

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/danielpaulus/go-ios/ios"
)

// muxDialer opens a connection to usbmuxd.
type muxDialer func() (*ios.UsbMuxConnection, error)

// usbPorts reaches the ports of a device over usbmux.
type usbPorts struct {
	dial   muxDialer
	listen func(network, address string) (net.Listener, error)
}

// Probe checks that something listens on a port of the device.
func (p usbPorts) Probe(entry ios.DeviceEntry, port uint16) error {
	mux, err := p.dial()
	if err != nil {
		return fmt.Errorf("usbmux: %w", err)
	}
	defer mux.Close()
	if err := mux.Connect(entry.DeviceID, port); err != nil {
		return fmt.Errorf("nothing listens on port %d of the device (%v)", port, err)
	}
	return nil
}

// Forward listens on a local port and forwards its connections to a port of
// the device.
func (p usbPorts) Forward(entry ios.DeviceEntry, phonePort uint16) (Forwarder, error) {
	ln, err := p.listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	f := &portForward{ln: ln, ctx: ctx, cancel: cancel, dial: p.dial}
	go f.accept(entry.DeviceID, phonePort)
	return f, nil
}

// portForward forwards connections to a local port to a port on the device
// over usbmux. Unlike go-ios' forward package it can be closed.
type portForward struct {
	ln     net.Listener
	ctx    context.Context
	cancel context.CancelFunc
	dial   muxDialer
}

func (f *portForward) accept(deviceID int, phonePort uint16) {
	for {
		conn, err := f.ln.Accept()
		if err != nil {
			return
		}
		go f.proxy(conn, deviceID, phonePort)
	}
}

func (f *portForward) Addr() string { return f.ln.Addr().String() }

func (f *portForward) Close() {
	f.cancel()
	_ = f.ln.Close()
}

// proxy connects the local connection to the device port until either ends.
func (f *portForward) proxy(conn net.Conn, deviceID int, phonePort uint16) {
	defer closeQuietly(conn)
	mux, err := f.dial()
	if err != nil {
		return
	}
	if err := mux.Connect(deviceID, phonePort); err != nil {
		mux.Close()
		return
	}
	dc := mux.ReleaseDeviceConnection()
	defer closeQuietly(dc)
	f.pipe(conn, dc)
}

// pipe copies in both directions until one side ends or the forward is closed.
func (f *portForward) pipe(conn net.Conn, dc ios.DeviceConnectionInterface) {
	ctx, cancel := context.WithCancel(f.ctx)
	go func() {
		_, _ = io.Copy(conn, dc.Reader())
		cancel()
	}()
	go func() {
		_, _ = io.Copy(dc.Writer(), conn)
		cancel()
	}()
	<-ctx.Done()
}

func closeQuietly(c io.Closer) { _ = c.Close() }
