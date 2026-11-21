package runstore

import (
	"context"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/protocol"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

// RunWithProtocol loads the run with the protocol, its device, entries,
// performance and keys; the run has no protocols when the protocol is not
// one of its.
func (s *Store) RunWithProtocol(ctx context.Context, runID, protocolID uint) (*models.TestRun, error) {
	var found models.TestRun
	err := s.session(ctx).Preload("Protocols", "id = ?", protocolID).Preload("Protocols.Device").
		Preload("Protocols.Entries").Preload("Protocols.Performance").
		Preload("Protocols.Keys", func(db *gorm.DB) *gorm.DB { return db.Order("key") }).
		Preload("Log").Preload("AppBinary").Preload("Test").First(&found, runID).Error
	if err != nil {
		return nil, lookup(err, "run", apperr.NotFound("run %d not found", runID))
	}
	return &found, nil
}

// History lists the earlier protocols of the same test name on the same
// device that measured performance, newest first, among the runs of the
// app: not the run itself (its runner protocol shares the test name but has
// no measurements), not later runs, and not runners of earlier runs, which
// would count as zeros.
func (s *Store) History(ctx context.Context, appID uint, current *models.TestProtocol) ([]models.TestProtocol, error) {
	var history []models.TestProtocol
	err := s.session(ctx).
		Where("device_id = ? and test_name = ? and test_run_id < ?", current.DeviceID, current.TestName, current.TestRunID).
		Where("test_run_id in (select tr.id from test_runs tr where tr.test_id in "+appTests+")", appID).
		Where("exists (select 1 from protocol_performance_entries pe where pe.test_protocol_id = test_protocols.id and pe.deleted_at is null)").
		Order("id desc").Limit(protocol.HistoryLimit).Preload("Performance").Find(&history).Error
	return history, storeerr.Map(err, "protocol history")
}

// Protocol loads the protocol of the key if its run belongs to the test and
// the test to the app of the project.
func (s *Store) Protocol(ctx context.Context, key protocol.Key) (*models.TestProtocol, error) {
	var found models.TestProtocol
	apps := s.session(ctx).Session(&gorm.Session{NewDB: true}).Model(&models.App{}).Select("id").Where("project_id = ?", key.ProjectID)
	err := s.session(ctx).Joins("JOIN test_runs ON test_runs.id = test_protocols.test_run_id").
		Joins("JOIN tests ON tests.id = test_runs.test_id").
		Where("test_protocols.id = ? AND test_protocols.test_run_id = ? AND test_runs.test_id = ? AND tests.app_id = ? AND tests.app_id IN (?)",
			key.ProtocolID, key.RunID, key.TestID, key.AppID, apps).
		First(&found).Error
	if err != nil {
		return nil, lookup(err, "protocol", apperr.NotFound("protocol %d not found in run %d", key.ProtocolID, key.RunID))
	}
	return &found, nil
}
