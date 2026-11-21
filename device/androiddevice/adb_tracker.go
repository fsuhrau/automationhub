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

	"github.com/fsuhrau/automationhub/device"
)

// DefaultADBServerAddr is where the adb server listens (adb start-server).
const DefaultADBServerAddr = "127.0.0.1:5037"

const maxBackoff = 30 * time.Second

const (
	trackDevicesLong  = "host:track-devices-l" // same format as "adb devices -l"
	trackDevicesShort = "host:track-devices"   // "serial\tstate", older adb servers
)

// errADBFailed is an error answer ("FAIL") from the adb server.
var errADBFailed = errors.New("adb server refused the request")

// adbSubscription is a track-devices request to the adb server.
type adbSubscription struct {
	addr    string
	request string
}

// trackDevices subscribes to device list changes of the adb server and calls
// onList with the complete list once right away and again on every change.
// It returns when ctx is done or the connection to the server breaks.
func trackDevices(ctx context.Context, sub adbSubscription, onList func([]adbDeviceEntry)) error {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", sub.addr)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	// unblock the read loop when ctx is cancelled
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()

	r := bufio.NewReader(conn)
	if err := requestTracking(conn, r, sub.request); err != nil {
		return err
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

// requestTracking sends the request and checks the answer of the server.
func requestTracking(conn net.Conn, r *bufio.Reader, request string) error {
	if _, err := fmt.Fprintf(conn, "%04x%s", len(request), request); err != nil {
		return err
	}
	status := make([]byte, 4)
	if _, err := io.ReadFull(r, status); err != nil {
		return err
	}
	switch string(status) {
	case "OKAY":
		return nil
	case "FAIL":
		msg, _ := readADBMessage(r)
		return fmt.Errorf("%w: %s", errADBFailed, msg)
	}
	return fmt.Errorf("unexpected adb server answer %q", status)
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

// adbWatcher keeps a track-devices subscription open and reconnects with
// backoff when the adb server goes away.
type adbWatcher struct {
	addr        string
	clock       device.Clock
	startServer func() error // restarts the adb server, may be nil
	onList      func([]adbDeviceEntry)
	// connected is called with true while a subscription delivers updates and
	// false when it broke
	connected func(bool)
}

func (w adbWatcher) run(ctx context.Context) {
	request := trackDevicesLong
	backoff := time.Second
	for ctx.Err() == nil {
		err := trackDevices(ctx, adbSubscription{addr: w.addr, request: request}, func(entries []adbDeviceEntry) {
			backoff = time.Second
			w.connected(true)
			w.onList(entries)
		})
		w.connected(false)
		if ctx.Err() != nil {
			return
		}
		if errors.Is(err, errADBFailed) && request == trackDevicesLong {
			// adb servers before 1.0.40 only know the short variant
			request = trackDevicesShort
			continue
		}
		if !w.wait(ctx, backoff) {
			return
		}
		backoff = nextBackoff(backoff)
	}
}

// wait restarts the adb server and sleeps; false when ctx ended meanwhile.
func (w adbWatcher) wait(ctx context.Context, backoff time.Duration) bool {
	if w.startServer != nil {
		_ = w.startServer()
	}
	select {
	case <-ctx.Done():
		return false
	case <-w.clock.After(backoff):
		return true
	}
}

func nextBackoff(backoff time.Duration) time.Duration {
	if backoff < maxBackoff {
		return backoff * 2
	}
	return backoff
}
