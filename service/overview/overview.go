package overview

import (
	"context"

	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

// LatestRuns is how many of the newest runs the overview lists.
const LatestRuns = 10

// Run is a run on the project overview, with its test and app.
type Run struct {
	testsvc.RunSummary
	TestID   uint   `json:"testId"`
	TestName string `json:"testName"`
	AppID    uint   `json:"appId"`
}

// Overview sums up a project.
type Overview struct {
	Apps          int64 `json:"apps"`
	Tests         int   `json:"tests"`
	Bundles       int64 `json:"bundles"`
	BundleStorage int64 `json:"bundleStorage"`
	// tests whose latest run failed, was unstable or didn't run
	Failing []Run `json:"failing"`
	// latest runs that are still running
	Running []Run `json:"running"`
	// the latest runs of the project's tests, newest first
	Latest []Run `json:"latest"`
}

// Get sums up the project: what it has, which tests need attention, what
// runs now and what ran last.
func (s *Service) Get(ctx context.Context, projectID uint) (*Overview, error) {
	totals, err := s.deps.Repo.Totals(ctx, projectID)
	if err != nil {
		return nil, err
	}
	tests, err := s.deps.Repo.Tests(ctx, projectID)
	if err != nil {
		return nil, err
	}
	o := &Overview{
		Apps: totals.Apps, Tests: len(tests), Bundles: totals.Bundles, BundleStorage: totals.BundleStorage,
		Failing: []Run{}, Running: []Run{}, Latest: []Run{},
	}
	if len(tests) == 0 {
		return o, nil
	}
	if err := s.attention(ctx, o, tests); err != nil {
		return nil, err
	}
	if err := s.latest(ctx, o, tests); err != nil {
		return nil, err
	}
	return o, nil
}

// attention sorts the latest run of each test into failing and running.
func (s *Service) attention(ctx context.Context, o *Overview, tests []models.Test) error {
	last, err := s.deps.Repo.LastRuns(ctx, testIDs(tests))
	if err != nil {
		return err
	}
	summaries, err := s.deps.Summarizer.Summaries(ctx, last)
	if err != nil {
		return err
	}
	byTest := byID(tests)
	for i := range last {
		switch summaries[i].Status {
		case "failed", "unstable", "not_run":
			o.Failing = append(o.Failing, entry(byTest, &last[i], summaries[i]))
		case "running":
			o.Running = append(o.Running, entry(byTest, &last[i], summaries[i]))
		}
	}
	return nil
}

// latest lists the newest runs of all the project's tests.
func (s *Service) latest(ctx context.Context, o *Overview, tests []models.Test) error {
	newest, err := s.deps.Repo.LatestRuns(ctx, testIDs(tests), LatestRuns)
	if err != nil {
		return err
	}
	summaries, err := s.deps.Summarizer.Summaries(ctx, newest)
	if err != nil {
		return err
	}
	byTest := byID(tests)
	for i := range newest {
		o.Latest = append(o.Latest, entry(byTest, &newest[i], summaries[i]))
	}
	return nil
}

func testIDs(tests []models.Test) []uint {
	ids := make([]uint, len(tests))
	for i := range tests {
		ids[i] = tests[i].ID
	}
	return ids
}

func byID(tests []models.Test) map[uint]models.Test {
	byTest := make(map[uint]models.Test, len(tests))
	for i := range tests {
		byTest[tests[i].ID] = tests[i]
	}
	return byTest
}

func entry(byTest map[uint]models.Test, r *models.TestRun, summary testsvc.RunSummary) Run {
	t := byTest[r.TestID]
	return Run{RunSummary: summary, TestID: t.ID, TestName: t.Name, AppID: t.AppID}
}
