package test

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/storage/models"
)

// RecentRunsPerTest is how many runs the tests list shows per test.
const RecentRunsPerTest = 10

// ListEntry is one test in the tests list.
type ListEntry struct {
	models.Test
	// the latest runs, newest first
	RecentRuns []RunSummary `json:"recentRuns"`
}

// List lists the tests of an app with their latest runs. It does not load
// every run of every test, which the list never shows.
func (s *Service) List(ctx context.Context, sc Scope) ([]ListEntry, error) {
	tests, err := s.repo.Tests(ctx, sc)
	if err != nil {
		return nil, fmt.Errorf("list tests: %w", err)
	}
	list := make([]ListEntry, len(tests))
	for i := range tests {
		list[i].Test = tests[i]
		if list[i].RecentRuns, err = s.recentRuns(ctx, tests[i].ID); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (s *Service) recentRuns(ctx context.Context, testID uint) ([]RunSummary, error) {
	recent, err := s.repo.RecentRuns(ctx, testID, RecentRunsPerTest)
	if err != nil {
		return nil, fmt.Errorf("recent runs of test %d: %w", testID, err)
	}
	return s.Summaries(ctx, recent)
}

// Get loads one test of the app; a test of another app or project is not
// found.
func (s *Service) Get(ctx context.Context, ref TestRef) (*models.Test, error) {
	return s.repo.Test(ctx, ref)
}
