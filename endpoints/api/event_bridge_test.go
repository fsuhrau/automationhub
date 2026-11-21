package api

import (
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/storage/models"
)

type bridgeProbe struct{ ID uint }

func TestBridgeEventPublishesChannelAndContent(t *testing.T) {
	tests := []struct {
		name        string
		channel     func(bridgeProbe) string
		wantChannel string
	}{
		{"fixed", fixedChannel[bridgeProbe]("probe"), "probe"},
		{"id", idChannel("probe_%d", func(p bridgeProbe) uint { return p.ID }), "probe_5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := events.NewBus[bridgeProbe]()
			r := make(eventRecorder, 1)
			bus.Register(newBridge(r, tt.channel, payloadContent))

			bus.Trigger(bridgeProbe{ID: 5})

			select {
			case e := <-r:
				if e.Channel != tt.wantChannel || e.Content != (bridgeProbe{ID: 5}) {
					t.Fatalf("event %+v", e)
				}
			case <-time.After(time.Second):
				t.Fatal("nothing published")
			}
		})
	}
}

func TestProtocolForUIWithoutDevice(t *testing.T) {
	p := &models.TestProtocol{TestName: "x"}
	got := protocolForUI(p)
	if got == p || got.TestName != "x" || got.Device != nil {
		t.Fatalf("copy %+v", got)
	}
}

var _ sse.Publisher = eventRecorder(nil)
