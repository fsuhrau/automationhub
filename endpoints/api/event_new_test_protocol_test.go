package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/storage/models"
)

// driver stands in for a device driver, which keeps its device config
type driver struct {
	Config *models.Device `json:"config"`
}

// During a run the protocol's device carries its live driver, which points
// back to the device. The event must still encode, or the run page keeps
// waiting for the first test.
func TestNewProtocolEventEncodesWithLiveDevice(t *testing.T) {
	dev := &models.Device{DeviceIdentifier: "phone", Name: "Pixel"}
	dev.Dev = &driver{Config: dev}
	dev.Connection = &driver{Config: dev}
	protocol := &models.TestProtocol{TestRunID: 7, TestName: "Suite/Test", Device: dev}
	protocol.ID = 3

	r := make(eventRecorder, 1)
	RegisterNewTestProtocolListener(r)
	events.NewTestProtocol.Trigger(events.NewTestProtocolPayload{TestRunID: 7, Protocol: protocol})

	var e sse.Event
	select {
	case e = <-r:
	case <-time.After(time.Second):
		t.Fatal("no protocol event")
	}
	if e.Channel != "test_run_7_protocol" {
		t.Fatalf("channel %q", e.Channel)
	}
	data, err := json.Marshal(e.Content)
	if err != nil {
		t.Fatalf("protocol event doesn't encode: %v", err)
	}
	var got struct {
		ID     uint `json:"id"`
		Device struct {
			Name string `json:"name"`
		} `json:"device"`
	}
	if err := json.Unmarshal(data, &got); err != nil || got.ID != 3 || got.Device.Name != "Pixel" {
		t.Fatalf("event %s (%v)", data, err)
	}
	if dev.Dev == nil {
		t.Error("the run's device lost its driver")
	}
}
