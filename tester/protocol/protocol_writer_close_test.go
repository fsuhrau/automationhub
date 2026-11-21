package protocol

import (
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestResultCountsAdd(t *testing.T) {
	tests := []struct {
		name  string
		state models.TestResultState
		want  resultCounts
	}{
		{"open is no success", models.TestResultOpen, resultCounts{failed: 1}},
		{"success", models.TestResultSuccess, resultCounts{succeeded: 1}},
		{"unstable", models.TestResultUnstable, resultCounts{unstable: 1}},
		{"failed", models.TestResultFailed, resultCounts{failed: 1}},
		{"unknown state is no success", models.TestResultState(99), resultCounts{failed: 1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got resultCounts
			got.add(tc.state)
			if got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

// closeWithPassed closes a writer holding one protocol per entry of passed and
// returns the payload of the finished event.
func closeWithPassed(t *testing.T, passed ...bool) events.TestRunFinishedPayload {
	t.Helper()
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
	r := make(finishedRecorder, 4)
	events.TestRunFinished.Register(r)

	w := NewProtocolWriter(db, RunTarget{ProjectID: "p", AppID: 1, TestName: "t", Run: &run})
	for _, ok := range passed {
		p, err := w.NewProtocol(dev, "Suite/A")
		if err != nil {
			t.Fatal(err)
		}
		p.Writer.Passed(ok)
	}
	w.Close()
	for {
		select {
		case e := <-r:
			if e.TestRunID == run.ID {
				return e
			}
		case <-time.After(time.Second):
			t.Fatal("no finished event")
		}
	}
}

// Close skipped the first protocol, so a run with a single test counted none.
func TestCloseCountsTheFirstProtocol(t *testing.T) {
	tests := []struct {
		name        string
		passed      []bool
		wantOK      int
		wantFailed  int
		wantSuccess bool
	}{
		{"single passed", []bool{true}, 1, 0, true},
		{"single failed", []bool{false}, 0, 1, false},
		{"first failed of two", []bool{false, true}, 1, 1, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := closeWithPassed(t, tc.passed...)
			if e.Succeeded != tc.wantOK || e.Failed != tc.wantFailed || e.Success != tc.wantSuccess {
				t.Fatalf("got succeeded=%d failed=%d success=%v", e.Succeeded, e.Failed, e.Success)
			}
		})
	}
}
