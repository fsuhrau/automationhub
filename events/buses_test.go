package events

import (
	"testing"
	"time"
)

func checkWired[T any](t *testing.T, bus *Bus[T], payload T, equal func(a, b T) bool) {
	t.Helper()
	got := make(chan T, 1)
	bus.Register(HandlerFunc[T](func(p T) { got <- p }))
	bus.Trigger(payload)
	select {
	case p := <-got:
		if !equal(p, payload) {
			t.Fatalf("payload = %+v, want %+v", p, payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler not called")
	}
}

func TestEveryBusVariableIsWired(t *testing.T) {
	tests := map[string]func(*testing.T){
		"AppCreated": func(t *testing.T) {
			checkWired(t, AppCreated, AppCreatedPayload{AppID: 1}, func(a, b AppCreatedPayload) bool { return a == b })
		},
		"DeviceListChanged": func(t *testing.T) {
			checkWired(t, DeviceListChanged, DeviceListChangedPayload{DeviceID: 1}, func(a, b DeviceListChangedPayload) bool { return a == b })
		},
		"DeviceLockChanged": func(t *testing.T) {
			checkWired(t, DeviceLockChanged, DeviceLockChangedPayload{DeviceIdentifier: "x"}, func(a, b DeviceLockChangedPayload) bool { return a == b })
		},
		"DeviceStatusChanged": func(t *testing.T) {
			checkWired(t, DeviceStatusChanged, DeviceStatusChangedPayload{DeviceID: 1}, func(a, b DeviceStatusChangedPayload) bool { return a == b })
		},
		"NewTestLogEntry": func(t *testing.T) {
			checkWired(t, NewTestLogEntry, NewTestLogEntryPayload{TestRunID: 1}, func(a, b NewTestLogEntryPayload) bool { return a == b })
		},
		"NewTestProtocol": func(t *testing.T) {
			checkWired(t, NewTestProtocol, NewTestProtocolPayload{TestRunID: 1}, func(a, b NewTestProtocolPayload) bool { return a == b })
		},
		"NewTestProtocolLog": func(t *testing.T) {
			checkWired(t, NewTestProtocolLog, NewTestProtocolLogPayload{TestProtocolID: 1}, func(a, b NewTestProtocolLogPayload) bool { return a == b })
		},
		"NodeConnected": func(t *testing.T) {
			checkWired(t, NodeConnected, NodeConnectedPayload{Hostname: "h"}, func(a, b NodeConnectedPayload) bool { return a == b })
		},
		"NodeDevicesChanged": func(t *testing.T) {
			checkWired(t, NodeDevicesChanged, NodeDevicesChangedPayload{SentAt: 5}, func(a, b NodeDevicesChangedPayload) bool { return a.SentAt == b.SentAt })
		},
		"NodeDisconnected": func(t *testing.T) {
			checkWired(t, NodeDisconnected, NodeDisconnectedPayload{NodeIdentifier: "n"}, func(a, b NodeDisconnectedPayload) bool { return a == b })
		},
		"ProtocolKey": func(t *testing.T) {
			checkWired(t, ProtocolKey, ProtocolKeyPayload{TestRunID: 1}, func(a, b ProtocolKeyPayload) bool { return a == b })
		},
		"TestCreated": func(t *testing.T) {
			checkWired(t, TestCreated, TestCreatedPayload{TestID: 1}, func(a, b TestCreatedPayload) bool { return a == b })
		},
		"TestRunFinished": func(t *testing.T) {
			checkWired(t, TestRunFinished, TestRunFinishedPayload{TestRunID: 1}, func(a, b TestRunFinishedPayload) bool { return a == b })
		},
		"TestRunStarted": func(t *testing.T) {
			checkWired(t, TestRunStarted, TestRunStartedPayload{TestRunID: 1}, func(a, b TestRunStartedPayload) bool { return a == b })
		},
	}
	for name, run := range tests {
		t.Run(name, run)
	}
}
