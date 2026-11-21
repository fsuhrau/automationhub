package protocol_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/protocol"
	"github.com/fsuhrau/automationhub/storage/models"
)

func device(id uint) *uint { return &id }

func candidate() models.TestProtocol {
	return models.TestProtocol{
		Model: models.Model{ID: 9}, TestRunID: 7, DeviceID: device(3), StartedAt: time.Unix(100, 0),
		TestResult: models.TestResultSuccess, AvgFPS: 60, AvgTriangles: 5,
		Device: &models.Device{Name: "Pixel"},
		TestRun: &models.TestRun{TestID: 2, Parameter: "Level=1", AppBinary: &models.AppBinary{
			Model: models.Model{ID: 6}, Name: "Game", Version: "1.0", Platform: "android", Tags: "release",
		}},
	}
}

func TestBaselinesListsTheCandidatesAndTheTags(t *testing.T) {
	e := newEnv()
	e.repo.candidates = []models.TestProtocol{candidate(), {Model: models.Model{ID: 10}, TestRun: &models.TestRun{TestID: 4}}, {Model: models.Model{ID: 11}}}
	e.repo.tags = []string{"release,store", "nightly, release", ""}

	got, err := e.svc.Baselines(bg(), key(), protocol.Filter{Tag: " release ", Query: " #Candidate ", AnyDevice: true, Limit: 7})

	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(got.Tags) != "[nightly release store]" {
		t.Errorf("tags %v", got.Tags)
	}
	first := got.Candidates[0]
	if first.ProtocolID != 9 || first.TestRunID != 7 || first.TestID != 2 || first.Device != "device Pixel" ||
		first.Parameter != "Level=1" || first.FPS != 60 || first.Triangles != 5 || first.Bundle == nil || first.Bundle.Version != "1.0" {
		t.Errorf("candidate %+v", first)
	}
	if noBundle := got.Candidates[1]; noBundle.Bundle != nil || noBundle.TestID != 4 {
		t.Errorf("candidate of a run without bundle %+v", noBundle)
	}
	if bare := got.Candidates[2]; bare.Bundle != nil || bare.Device != "" || bare.TestID != 0 {
		t.Errorf("candidate without relations %+v", bare)
	}
	q := e.repo.candidateQuery
	if q.AppID != 2 || q.Filter.Tag != "release" || q.Filter.Query != "#candidate" || !q.Filter.AnyDevice || q.Filter.Limit != 7 {
		t.Errorf("candidate query %+v", q)
	}
}

func TestBaselinesWithoutBundleTagsHasAnEmptyTagList(t *testing.T) {
	e := newEnv()

	got, err := e.svc.Baselines(bg(), key(), protocol.Filter{})

	if err != nil || got.Tags == nil || len(got.Tags) != 0 || got.Candidates == nil {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestBaselineLoadsTheBaselineWithItsCheckpoints(t *testing.T) {
	e := newEnv()
	e.repo.baseline = candidate()
	e.repo.baseline.Performance = []models.ProtocolPerformanceEntry{{Checkpoint: "start"}, {Checkpoint: "end"}}
	e.repo.protocol = models.TestProtocol{TestName: "S/A", DeviceID: device(3)}

	got, err := e.svc.Baseline(bg(), key(), protocol.Choice{RunID: 7, DeviceID: device(4)})

	if err != nil || got.ProtocolID != 9 || len(got.Performance) != 2 {
		t.Fatalf("got %+v, %v", got, err)
	}
	q := e.repo.baselineQuery
	if q.AppID != 2 || q.RunID != 7 || *q.DeviceID != 4 || q.Current.TestName != "S/A" {
		t.Errorf("baseline query %+v", q)
	}
}

func TestBaselinesAndBaselineErrors(t *testing.T) {
	internal := apperr.Internal("db")
	tests := []struct {
		name string
		fail string
		call func(*env) error
	}{
		{"baselines protocol", "Protocol", func(e *env) error { _, err := e.svc.Baselines(bg(), key(), protocol.Filter{}); return err }},
		{"baselines candidates", "Candidates", func(e *env) error { _, err := e.svc.Baselines(bg(), key(), protocol.Filter{}); return err }},
		{"baselines tags", "BundleTags", func(e *env) error { _, err := e.svc.Baselines(bg(), key(), protocol.Filter{}); return err }},
		{"baseline protocol", "Protocol", func(e *env) error { _, err := e.svc.Baseline(bg(), key(), protocol.Choice{RunID: 7}); return err }},
		{"baseline load", "Baseline", func(e *env) error { _, err := e.svc.Baseline(bg(), key(), protocol.Choice{RunID: 7}); return err }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv()
			e.repo.fail[tc.fail] = internal

			assertKind(t, tc.call(e), apperr.KindInternal)
		})
	}
}
