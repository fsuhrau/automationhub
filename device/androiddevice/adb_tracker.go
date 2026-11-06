package androiddevice

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// DefaultADBServerAddr is where the adb server listens (adb start-server).
const DefaultADBServerAddr = "127.0.0.1:5037"

const (
	trackDevicesLong  = "host:track-devices-l" // same format as "adb devices -l"
	trackDevicesShort = "host:track-devices"   // "serial\tstate", older adb servers
)

// errADBFailed is an error answer ("FAIL") from the adb server.
var errADBFailed = errors.New("adb server refused the request")

// trackDevices subscribes to device list changes of the adb server and calls
// onList with the complete list once right away and again on every change.
// It returns when ctx is done or the connection to the server breaks.
func trackDevices(ctx context.Context, addr, request string, onList func([]adbDeviceEntry)) error {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// unblock the read loop when ctx is cancelled
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()

	if _, err := fmt.Fprintf(conn, "%04x%s", len(request), request); err != nil {
		return err
	}
	r := bufio.NewReader(conn)
	status := make([]byte, 4)
	if _, err := io.ReadFull(r, status); err != nil {
		return err
	}
	switch string(status) {
	case "OKAY":
	case "FAIL":
		msg, _ := readADBMessage(r)
		return fmt.Errorf("%w: %s", errADBFailed, msg)
	default:
		return fmt.Errorf("unexpected adb server answer %q", status)
	}

	for {
		payload, err := readADBMessage(r)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		onList(parseADBDeviceList(payload))
	}
}

// readADBMessage reads one length-prefixed (4 hex digits) adb message.
func readADBMessage(r *bufio.Reader) (string, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(r, header); err != nil {
		return "", err
	}
	n, err := strconv.ParseUint(string(header), 16, 32)
	if err != nil {
		return "", fmt.Errorf("invalid adb message length %q", header)
	}
	payload := make([]byte, n)
	if _, err := io.ReadFull(r, payload); err != nil {
		return "", err
	}
	return string(payload), nil
}

// parseADBDeviceList parses the device list of "adb devices -l" or of a
// track-devices message.
func parseADBDeviceList(text string) []adbDeviceEntry {
	var entries []adbDeviceEntry
	for _, line := range strings.Split(text, "\n") {
		if entry, ok := parseADBDeviceLine(strings.TrimRight(line, "\r")); ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

// watchADBDevices keeps a track-devices subscription open and reconnects
// with backoff when the adb server goes away. connected is called with true
// while a subscription delivers updates and false when it broke.
func watchADBDevices(ctx context.Context, addr string, startServer func() error, onList func([]adbDeviceEntry), connected func(bool)) {
	request := trackDevicesLong
	backoff := time.Second
	for ctx.Err() == nil {
		err := trackDevices(ctx, addr, request, func(entries []adbDeviceEntry) {
			backoff = time.Second
			connected(true)
			onList(entries)
		})
		connected(false)
		if ctx.Err() != nil {
			return
		}
		if errors.Is(err, errADBFailed) && request == trackDevicesLong {
			// adb servers before 1.0.40 only know the short variant
			request = trackDevicesShort
			continue
		}
		if startServer != nil {
			_ = startServer()
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
