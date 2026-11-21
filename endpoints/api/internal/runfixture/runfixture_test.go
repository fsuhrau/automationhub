package runfixture_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/endpoints/api/internal/runfixture"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestTheWorldHasTwoProjectsWithOneTestAndRunEach(t *testing.T) {
	w := runfixture.New(t)

	if w.Foreign.ID >= w.Mine.ID || w.ForeignRun.ID >= w.Run1.ID || w.Run1.ID >= w.Run2.ID {
		t.Errorf("the foreign ids must be the lower ones: tests %d/%d, runs %d/%d/%d", w.Foreign.ID, w.Mine.ID, w.ForeignRun.ID, w.Run1.ID, w.Run2.ID)
	}
	var statuses, performance int64
	w.DB.Model(&models.TestRunDeviceStatus{}).Count(&statuses)
	w.DB.Model(&models.ProtocolPerformanceEntry{}).Count(&performance)
	if statuses != 3 || performance != 3 {
		t.Errorf("%d device statuses and %d performance entries, want 3 each", statuses, performance)
	}
	if w.Proto1.AvgFPS != 10 || w.Proto2.AvgFPS != 0 || w.ForeProto.AvgFPS != 100 {
		t.Errorf("fps %v %v %v", w.Proto1.AvgFPS, w.Proto2.AvgFPS, w.ForeProto.AvgFPS)
	}
}

func TestPathIsBelowTheAppOfTheProjectMine(t *testing.T) {
	w := runfixture.New(t)

	rec := w.Do("GET", w.Path("/test/%d/runs", w.Mine.ID), nil)

	if want := fmt.Sprintf("/api/mine/app/%d/test/1", w.MineApp.ID); w.Path("/test/1") != want {
		t.Errorf("path %q, want %q", w.Path("/test/1"), want)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
}
