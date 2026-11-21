package runstore

import (
	"context"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

// ReportRun loads a run with its test, bundle and, by id, its protocols
// with their device and keys.
func (s *Store) ReportRun(ctx context.Context, runID uint) (*models.TestRun, error) {
	var found models.TestRun
	err := s.session(ctx).Preload("Test").Preload("AppBinary").
		Preload("Protocols", func(db *gorm.DB) *gorm.DB { return db.Order("id") }).
		Preload("Protocols.Device").Preload("Protocols.Keys").
		First(&found, runID).Error
	if err != nil {
		return nil, lookup(err, "run", apperr.NotFound("not found: test run %d", runID))
	}
	return &found, nil
}

// LogTail lists the newest log entries of the run, newest first.
func (s *Store) LogTail(ctx context.Context, runID uint, limit int) ([]models.TestRunLogEntry, error) {
	var entries []models.TestRunLogEntry
	err := s.session(ctx).Where("test_run_id = ?", runID).Order("id desc").Limit(limit).Find(&entries).Error
	return entries, storeerr.Map(err, "run log")
}

// ProtocolProgress counts the protocols of the run and those that ended.
func (s *Store) ProtocolProgress(ctx context.Context, runID uint) (run.Progress, error) {
	protocols := func() *gorm.DB {
		return s.session(ctx).Model(&models.TestProtocol{}).Where("test_run_id = ?", runID)
	}
	var progress run.Progress
	if err := protocols().Count(&progress.Total).Error; err != nil {
		return run.Progress{}, storeerr.Map(err, "protocol progress")
	}
	err := protocols().Where("ended_at IS NOT NULL").Count(&progress.Ended).Error
	return progress, storeerr.Map(err, "protocol progress")
}
