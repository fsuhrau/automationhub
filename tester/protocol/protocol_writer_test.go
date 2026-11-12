package protocol

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type protocolRecorder chan events.NewTestProtocolPayload

func (r protocolRecorder) Handle(p events.NewTestProtocolPayload) { r <- p }

// A test that reported no performance samples (e.g. the runner's protocol of a
// Unity test) must still finish with values that encode as JSON; NaN averages
// kept the finished test from reaching the run page.
func TestClosedProtocolWithoutSamplesEncodes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Device{}, &models.TestRun{}, &models.TestProtocol{}, &models.ProtocolEntry{}, &models.ProtocolPerformanceEntry{}); err != nil {
		t.Fatal(err)
	}
	dev := models.Device{DeviceIdentifier: "phone"}
	db.Create(&dev)
	run := models.TestRun{}
	db.Create(&run)

	r := make(protocolRecorder, 4)
	events.NewTestProtocol.Register(r)
	p, err := NewProtocolWriter(db, "p", 1, "t", &run).NewProtocol(dev, "Suite/A")
	if err != nil {
		t.Fatal(err)
	}
	p.Close()

	for i := 0; i < 2; i++ {
		select {
		case e := <-r:
			if e.TestRunID != run.ID {
				continue
			}
			pr := e.Protocol.(*models.TestProtocol)
			if pr.EndedAt == nil {
				continue
			}
			if _, err := json.Marshal(pr); err != nil {
				t.Fatalf("finished protocol doesn't encode: %v", err)
			}
			return
		case <-time.After(time.Second):
			t.Fatal("no finished protocol event")
		}
	}
	t.Fatal("no finished protocol event")
}
