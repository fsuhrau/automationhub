package app

import (
	"context"

	"github.com/fsuhrau/automationhub/storage/models"
)

// detail adds the runs that used the bundle to it.
func (s *Service) detail(ctx context.Context, b *models.AppBinary) (*BundleDetail, error) {
	count, err := s.repo.RunCount(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	testRuns, err := s.repo.RecentRuns(ctx, b.ID, bundleRunsLimit)
	if err != nil {
		return nil, err
	}
	summaries, err := s.summarize(testRuns)
	if err != nil {
		return nil, err
	}
	detail := &BundleDetail{AppBinary: *b, RunCount: count, Runs: make([]BundleRun, len(testRuns))}
	for i, r := range testRuns {
		detail.Runs[i] = BundleRun{RunSummary: summaries[i], TestID: r.TestID}
		if r.Test != nil {
			detail.Runs[i].TestName = r.Test.Name
		}
	}
	return detail, nil
}

// inUse says whether a test run of this master runs with the bundle. A
// running run is one of the latest that used the bundle.
func (s *Service) inUse(ctx context.Context, bundleID uint) (bool, error) {
	ids, err := s.repo.RecentRunIDs(ctx, bundleID, 100)
	if err != nil {
		return false, err
	}
	for _, id := range ids {
		if s.running(id) {
			return true, nil
		}
	}
	return false, nil
}
