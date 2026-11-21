package appstore

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
)

// RunCounts counts the test runs that used each of the bundles; a bundle no
// run used is not in the result.
func (s *Store) RunCounts(ctx context.Context, bundleIDs []uint) (map[uint]int64, error) {
	counts := map[uint]int64{}
	if len(bundleIDs) == 0 {
		return counts, nil
	}
	var rows []struct {
		AppBinaryID uint
		Runs        int64
	}
	err := s.tx(ctx).Model(&models.TestRun{}).Select("app_binary_id, count(*) as runs").
		Where("app_binary_id IN ?", bundleIDs).Group("app_binary_id").Scan(&rows).Error
	for _, r := range rows {
		counts[r.AppBinaryID] = r.Runs
	}
	return counts, storeerr.Map(err, "run counts")
}

// RunCount counts the test runs that used the bundle.
func (s *Store) RunCount(ctx context.Context, bundleID uint) (int64, error) {
	var n int64
	err := s.tx(ctx).Model(&models.TestRun{}).Where("app_binary_id = ?", bundleID).Count(&n).Error
	return n, storeerr.Map(err, fmt.Sprintf("runs of bundle %d", bundleID))
}

// RecentRuns lists the newest runs that used the bundle, with their test.
func (s *Store) RecentRuns(ctx context.Context, bundleID uint, limit int) ([]models.TestRun, error) {
	var found []models.TestRun
	err := s.tx(ctx).Preload("Test").Where("app_binary_id = ?", bundleID).Order("id desc").Limit(limit).Find(&found).Error
	return found, storeerr.Map(err, fmt.Sprintf("runs of bundle %d", bundleID))
}

// RecentRunIDs lists the ids of the newest runs that used the bundle.
func (s *Store) RecentRunIDs(ctx context.Context, bundleID uint, limit int) ([]uint, error) {
	var ids []uint
	err := s.tx(ctx).Model(&models.TestRun{}).Where("app_binary_id = ?", bundleID).
		Order("id desc").Limit(limit).Pluck("id", &ids).Error
	return ids, storeerr.Map(err, fmt.Sprintf("runs of bundle %d", bundleID))
}
