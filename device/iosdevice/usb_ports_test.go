package iosdevice

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
)

// fakeUsbmuxd answers the Connect message of a usbmux connection with number
// (0 is success) and then echoes what it receives like a service on the device.
func fakeUsbmuxd(t *testing.T, number int) muxDialer {
	t.Helper()
	return func() (*ios.UsbMuxConnection, error) {
		client, server := net.Pipe()
		go serveMux(server, number)
		return ios.NewUsbMuxConnection(ios.NewDeviceConnectionWithConn(client)), nil
	}
}

func serveMux(conn net.Conn, number int) {
	defer func() { _ = conn.Close() }()
	var header ios.UsbMuxHeader
	if binary.Read(conn, binary.LittleEndian, &header) != nil {
		return
	}
	if _, err := io.CopyN(io.Discard, conn, int64(header.Length-16)); err != nil {
		return
	}
	payload := ios.ToPlistBytes(map[string]any{"MessageType": "Result", "Number": number})
	reply := ios.UsbMuxHeader{Length: 16 + uint32(len(payload)), Version: 1, Request: 8, Tag: header.Tag}
	if binary.Write(conn, binary.LittleEndian, reply) != nil {
		return
	}
	if _, err := conn.Write(payload); err != nil || number != 0 {
		return
	}
	_, _ = io.Copy(conn, conn) // the device service echoes
}

func newPorts(dial muxDialer) usbPorts { return usbPorts{dial: dial, listen: net.Listen} }

func TestForwardReportsAListenError(t *testing.T) {
	ports := usbPorts{listen: func(string, string) (net.Listener, error) { return nil, errors.New("no ports left") }}
	if _, err := ports.Forward(entryOf("U"), 8100); err == nil {
		t.Fatal("expected an error")
	}
}

func TestProbeFindsAListeningPort(t *testing.T) {
	if err := (newPorts(fakeUsbmuxd(t, 0))).Probe(entryOf("U"), 8100); err != nil {
		t.Fatal(err)
	}
}

func TestProbeReportsARefusedPort(t *testing.T) {
	err := newPorts(fakeUsbmuxd(t, 3)).Probe(entryOf("U"), 8100)
	if err == nil || !strings.Contains(err.Error(), "nothing listens on port 8100 of the device") {
		t.Fatalf("err = %v", err)
	}
}

func TestProbeReportsAMissingUsbmuxd(t *testing.T) {
	dial := func() (*ios.UsbMuxConnection, error) { return nil, errors.New("no socket") }
	if err := (newPorts(dial)).Probe(entryOf("U"), 8100); err == nil || !strings.HasPrefix(err.Error(), "usbmux: ") {
		t.Fatalf("err = %v", err)
	}
}

func TestForwardCarriesBytesToTheDevicePort(t *testing.T) {
	f, err := newPorts(fakeUsbmuxd(t, 0)).Forward(entryOf("U"), 8100)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	conn, err := net.Dial("tcp", f.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := io.WriteString(conn, "ping"); err != nil {
		t.Fatal(err)
	}
	got := make([]byte, 4)
	if _, err := io.ReadFull(conn, got); err != nil || string(got) != "ping" {
		t.Fatalf("echo = %q, %v", got, err)
	}
}

func TestForwardDropsConnectionsWhenTheDeviceRefuses(t *testing.T) {
	for name, dial := range map[string]muxDialer{
		"refused": fakeUsbmuxd(t, 3),
		"no mux":  func() (*ios.UsbMuxConnection, error) { return nil, errors.New("no socket") },
	} {
		f, err := newPorts(dial).Forward(entryOf("U"), 8100)
		if err != nil {
			t.Fatal(err)
		}
		conn, err := net.Dial("tcp", f.Addr())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := conn.Read(make([]byte, 1)); err == nil {
			t.Errorf("%s: the connection should end", name)
		}
		_ = conn.Close()
		f.Close()
	}
}

func TestForwardStopsListeningAfterClose(t *testing.T) {
	f, err := newPorts(fakeUsbmuxd(t, 0)).Forward(entryOf("U"), 8100)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if conn, err := net.Dial("tcp", f.Addr()); err == nil {
		_ = conn.Close()
		t.Fatal("the forward still listens after Close")
	}
}
