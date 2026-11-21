package teststore

import (
	"context"

	service "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
)

// RecentRuns lists the newest runs of a test, newest first.
func (s *Store) RecentRuns(ctx context.Context, testID uint, limit int) ([]models.TestRun, error) {
	var recent []models.TestRun
	err := s.session(ctx).Where("test_id = ?", testID).Order("id desc").Limit(limit).Find(&recent).Error
	return recent, storeerr.Map(err, "recent runs")
}

// RunResults counts the test protocols of the runs per result, leaving out
// the runner protocols that wrap them.
func (s *Store) RunResults(ctx context.Context, runIDs []uint) ([]service.ResultCount, error) {
	var results []service.ResultCount
	err := s.session(ctx).Model(&models.TestProtocol{}).
		Select("test_run_id, test_result, count(*) as n").
		Scopes(InRuns(runIDs), LeafProtocols).
		Group("test_run_id, test_result").Scan(&results).Error
	return results, storeerr.Map(err, "run results")
}

// RunLogErrors counts the error log entries of the runs.
func (s *Store) RunLogErrors(ctx context.Context, runIDs []uint) ([]service.LogErrorCount, error) {
	var logErrors []service.LogErrorCount
	err := s.session(ctx).Model(&models.TestRunLogEntry{}).Select("test_run_id, count(*) as n").
		Where("test_run_id IN ? AND level = ?", runIDs, "error").
		Group("test_run_id").Scan(&logErrors).Error
	return logErrors, storeerr.Map(err, "run log errors")
}
