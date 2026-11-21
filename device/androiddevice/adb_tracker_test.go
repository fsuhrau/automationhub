package androiddevice

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/internal/testkit"
)

// fakeADBServer speaks the adb host protocol for track-devices.
type fakeADBServer struct {
	ln       net.Listener
	mu       sync.Mutex
	requests []string
	// handle serves one accepted connection after the request was read
	handle func(conn net.Conn, request string)
}

func newFakeADBServer(t *testing.T, handle func(conn net.Conn, request string)) *fakeADBServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &fakeADBServer{ln: ln, handle: handle}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() { _ = conn.Close() }()
				r := bufio.NewReader(conn)
				header := make([]byte, 4)
				if _, err := io.ReadFull(r, header); err != nil {
					return
				}
				n, _ := strconv.ParseUint(string(header), 16, 32)
				body := make([]byte, n)
				if _, err := io.ReadFull(r, body); err != nil {
					return
				}
				s.mu.Lock()
				s.requests = append(s.requests, string(body))
				s.mu.Unlock()
				s.handle(conn, string(body))
			}()
		}
	}()
	t.Cleanup(func() { _ = ln.Close() })
	return s
}

func (s *fakeADBServer) addr() string { return s.ln.Addr().String() }

func (s *fakeADBServer) requestList() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.requests...)
}

func writeADBMessage(conn net.Conn, payload string) {
	_, _ = fmt.Fprintf(conn, "%04x%s", len(payload), payload)
}

func TestTrackDevicesDeliversEveryChange(t *testing.T) {
	updates := []string{
		"",
		"RZCTA1YRJ5W\tunauthorized usb:0-1.3 transport_id:2\n",
		"RZCTA1YRJ5W\tdevice usb:0-1.3 product:r9q2xeea model:SM_G990B2 device:r9q transport_id:3\nemulator-5554\tdevice product:sdk model:Pixel device:emu transport_id:4\n",
	}
	srv := newFakeADBServer(t, func(conn net.Conn, request string) {
		_, _ = conn.Write([]byte("OKAY"))
		for _, u := range updates {
			writeADBMessage(conn, u)
		}
		time.Sleep(time.Second)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var lists [][]adbDeviceEntry
	err := trackDevices(ctx, adbSubscription{addr: srv.addr(), request: trackDevicesLong}, func(entries []adbDeviceEntry) {
		lists = append(lists, entries)
		if len(lists) == len(updates) {
			cancel()
		}
	})
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if got := srv.requestList(); len(got) != 1 || got[0] != trackDevicesLong {
		t.Fatalf("unexpected requests %v", got)
	}
	if len(lists[0]) != 0 || lists[1][0].State != adbStateUnauthorized || len(lists[2]) != 2 || lists[2][0].Fields["model"] != "SM_G990B2" {
		t.Fatalf("unexpected lists %+v", lists)
	}
}

func TestWatchFallsBackToShortTrackDevices(t *testing.T) {
	srv := newFakeADBServer(t, func(conn net.Conn, request string) {
		if request == trackDevicesLong {
			_, _ = conn.Write([]byte("FAIL"))
			writeADBMessage(conn, "unknown host service")
			return
		}
		_, _ = conn.Write([]byte("OKAY"))
		writeADBMessage(conn, "emulator-5554\tdevice\n")
		time.Sleep(time.Second)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	got := make(chan []adbDeviceEntry, 1)
	w := adbWatcher{addr: srv.addr(), clock: &autoClock{}, onList: func(entries []adbDeviceEntry) {
		select {
		case got <- entries:
		default:
		}
	}, connected: func(bool) {}}
	go w.run(ctx)

	select {
	case entries := <-got:
		if len(entries) != 1 || entries[0].Serial != "emulator-5554" || entries[0].State != adbStateDevice {
			t.Fatalf("unexpected entries %+v", entries)
		}
	case <-ctx.Done():
		t.Fatal("no device list received")
	}
	if reqs := srv.requestList(); len(reqs) < 2 || reqs[0] != trackDevicesLong || reqs[1] != trackDevicesShort {
		t.Fatalf("expected fallback to %s, got %v", trackDevicesShort, reqs)
	}
}

func TestWatchReconnectsAfterServerRestart(t *testing.T) {
	var mu sync.Mutex
	connections := 0
	srv := newFakeADBServer(t, func(conn net.Conn, request string) {
		mu.Lock()
		connections++
		n := connections
		mu.Unlock()
		_, _ = conn.Write([]byte("OKAY"))
		writeADBMessage(conn, fmt.Sprintf("device-%d\tdevice\n", n))
		if n == 1 {
			return // the server goes away after the first update
		}
		time.Sleep(2 * time.Second)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	serials := make(chan string, 4)
	var states []bool
	var statesMu sync.Mutex
	w := adbWatcher{addr: srv.addr(), clock: &autoClock{}, startServer: func() error { return nil }, onList: func(entries []adbDeviceEntry) {
		serials <- entries[0].Serial
	}, connected: func(up bool) {
		statesMu.Lock()
		states = append(states, up)
		statesMu.Unlock()
	}}
	go w.run(ctx)

	for _, want := range []string{"device-1", "device-2"} {
		select {
		case got := <-serials:
			if got != want {
				t.Fatalf("expected %s, got %s", want, got)
			}
		case <-ctx.Done():
			t.Fatalf("no update %s (reconnect missing)", want)
		}
	}
	statesMu.Lock()
	defer statesMu.Unlock()
	if len(states) < 3 || !states[0] || states[1] || !states[2] {
		t.Fatalf("expected connected, disconnected, connected; got %v", states)
	}
}

func TestTrackDevicesErrors(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closedAddr := ln.Addr().String()
	_ = ln.Close()
	drop := newFakeADBServer(t, func(conn net.Conn, _ string) {}) // closes without an answer
	odd := newFakeADBServer(t, func(conn net.Conn, _ string) { _, _ = conn.Write([]byte("WHAT")) })
	truncated := newFakeADBServer(t, func(conn net.Conn, _ string) {
		_, _ = conn.Write([]byte("OKAYzzzz"))
	})
	tests := []struct {
		name string
		addr string
		want string
	}{
		{"nothing listens", closedAddr, "connect"},
		{"no answer", drop.addr(), "EOF"},
		{"unknown answer", odd.addr(), `unexpected adb server answer "WHAT"`},
		{"bad message length", truncated.addr(), "invalid adb message length"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := trackDevices(context.Background(), adbSubscription{addr: tt.addr, request: trackDevicesLong}, func([]adbDeviceEntry) {})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestWatchStopsWhileWaitingToReconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	clock := testkit.NewFakeClock(time.Unix(0, 0)) // After never fires
	started := make(chan struct{}, 1)
	w := adbWatcher{addr: "127.0.0.1:1", clock: clock, startServer: func() error {
		select {
		case started <- struct{}{}:
		default:
		}
		return nil
	}, onList: func([]adbDeviceEntry) {}, connected: func(bool) {}}
	done := make(chan struct{})
	go func() { w.run(ctx); close(done) }()
	<-started
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("the watcher did not stop")
	}
}

func TestNextBackoffDoublesUpToTheLimit(t *testing.T) {
	if got := nextBackoff(time.Second); got != 2*time.Second {
		t.Fatalf("got %v", got)
	}
	if got := nextBackoff(maxBackoff); got != maxBackoff {
		t.Fatalf("got %v", got)
	}
}

func TestRequestTrackingFailsOnAClosedConnection(t *testing.T) {
	client, server := net.Pipe()
	_ = server.Close()
	_ = client.Close()
	if err := requestTracking(client, bufio.NewReader(client), trackDevicesLong); err == nil {
		t.Fatal("expected a write error")
	}
}
