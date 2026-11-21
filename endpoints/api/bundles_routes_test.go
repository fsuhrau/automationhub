package api

import (
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestSummarizeBundleRunsGivesTheRunsTheirSummary(t *testing.T) {
	db := testkit.DB(t)
	run := testkit.TestRun(db)
	s := &Service{db: db}
	got, err := s.summarizeBundleRuns([]models.TestRun{*run})
	if err != nil || len(got) != 1 || got[0].ID != run.ID || got[0].Status != "empty" {
		t.Fatalf("summaries %+v, %v", got, err)
	}
}

func TestSummarizeBundleRunsReportsAFailingDatabase(t *testing.T) {
	db := testkit.DB(t)
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
	s := &Service{db: db}
	if _, err := s.summarizeBundleRuns([]models.TestRun{{Model: models.Model{ID: 1}}}); err == nil {
		t.Fatal("a closed database must fail")
	}
}
