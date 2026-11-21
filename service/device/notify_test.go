package device_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	svc "github.com/fsuhrau/automationhub/service/device"
)

type fakeIdentities map[string]uint

func (f fakeIdentities) IDByIdentifier(_ context.Context, identifier string) (uint, error) {
	if id, ok := f[identifier]; ok {
		return id, nil
	}
	return 0, errors.New("unknown")
}

type logRecorder struct{ lines []string }

func (l *logRecorder) Debugf(format string, args ...any) {
	l.lines = append(l.lines, fmt.Sprintf(format, args...))
}

func listenOn(bus *events.Bus[events.DeviceListChangedPayload]) chan events.DeviceListChangedPayload {
	got := make(chan events.DeviceListChangedPayload, 1)
	bus.Register(events.HandlerFunc[events.DeviceListChangedPayload](func(p events.DeviceListChangedPayload) { got <- p }))
	return got
}

func TestListNotifierAnnouncesAKnownDevice(t *testing.T) {
	bus := events.NewBus[events.DeviceListChangedPayload]()
	got := listenOn(bus)

	svc.NewListNotifier(fakeIdentities{"phone": 7}, bus, &logRecorder{}).DeviceChanged("phone")

	select {
	case p := <-got:
		if p.DeviceID != 7 {
			t.Fatalf("payload %+v", p)
		}
	case <-time.After(time.Second):
		t.Fatal("no event")
	}
}

func TestListNotifierIgnoresAnUnknownDevice(t *testing.T) {
	bus := events.NewBus[events.DeviceListChangedPayload]()
	got := listenOn(bus)
	log := &logRecorder{}

	svc.NewListNotifier(fakeIdentities{}, bus, log).DeviceChanged("nobody")

	select {
	case p := <-got:
		t.Fatalf("unexpected reload %+v", p)
	case <-time.After(50 * time.Millisecond):
	}
	if len(log.lines) != 1 {
		t.Fatalf("log %v", log.lines)
	}
}
