package protocol

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/internal/testkit"
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
	p, err := NewProtocolWriter(db, RunTarget{ProjectID: "p", AppID: 1, TestName: "t", Run: &run}).NewProtocol(dev, "Suite/A")
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

type finishedRecorder chan events.TestRunFinishedPayload

func (r finishedRecorder) Handle(p events.TestRunFinishedPayload) { r <- p }

// A run that couldn't run its tests (e.g. no device) failed, although no
// test failed; it used to be reported as a success.
func TestFailedRunIsNoSuccess(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.TestRun{}); err != nil {
		t.Fatal(err)
	}
	r := make(finishedRecorder, 4)
	events.TestRunFinished.Register(r)

	result := func(fail bool) bool {
		run := models.TestRun{}
		db.Create(&run)
		w := NewProtocolWriter(db, RunTarget{ProjectID: "p", AppID: 1, TestName: "t", Run: &run})
		if fail {
			w.Fail()
		}
		w.Close()
		for {
			select {
			case e := <-r:
				if e.TestRunID == run.ID {
					return e.Success
				}
			case <-time.After(time.Second):
				t.Fatal("no finished event")
			}
		}
	}
	if result(true) {
		t.Fatal("failed run reported as success")
	}
	if !result(false) {
		t.Fatal("run without failures reported as failed")
	}
}

// The devices of a run start their tests at once; every protocol they open
// is remembered, with no data race (run with -race).
func TestNewProtocolIsSafeForTheDevicesOfARun(t *testing.T) {
	db := testkit.DB(t)
	run := testkit.TestRun(db)
	w := NewProtocolWriter(db, RunTarget{ProjectID: "p", AppID: 1, TestName: "t", Run: run})
	const devices = 8
	var wg sync.WaitGroup

	for i := 0; i < devices; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := w.NewProtocol(models.Device{}, "Suite/A"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	if got := len(w.tracked()); got != devices {
		t.Fatalf("%d protocols tracked, want %d", got, devices)
	}
}
