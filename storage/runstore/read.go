package runstore

import (
	"context"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/run"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

// TestInScope tells that the test belongs to the app and the project.
func (s *Store) TestInScope(ctx context.Context, ref testsvc.TestRef) error {
	apps := s.session(ctx).Session(&gorm.Session{NewDB: true}).Model(&models.App{}).Select("id").Where("project_id = ?", ref.ProjectID)
	query := s.session(ctx).Model(&models.Test{}).
		Where("id = ? AND app_id = ? AND app_id IN (?)", ref.TestID, ref.AppID, apps)
	return require(query, "test", apperr.NotFound("test %d not found", ref.TestID))
}

// RunInTest tells that the run belongs to the test.
func (s *Store) RunInTest(ctx context.Context, testID, runID uint) error {
	query := s.session(ctx).Model(&models.TestRun{}).Where("id = ? AND test_id = ?", runID, testID)
	return require(query, "run", apperr.NotFound("run %d not found", runID))
}

// TestRuns pages the runs of a test, newest first.
func (s *Store) TestRuns(ctx context.Context, q run.PageQuery) ([]models.TestRun, error) {
	query := s.session(ctx).Where("test_id = ?", q.TestID)
	if q.Before > 0 {
		query = query.Where("id < ?", q.Before)
	}
	var found []models.TestRun
	err := query.Preload("AppBinary", func(db *gorm.DB) *gorm.DB { return db.Omit("info") }).
		Order("id desc").Limit(q.Limit).Find(&found).Error
	return found, storeerr.Map(err, "test runs")
}

// RunEnds lists the end of every ended protocol of the runs. The ends are
// read as columns and compared by the caller, since an aggregate of a time
// column comes back as text on SQLite.
func (s *Store) RunEnds(ctx context.Context, runIDs []uint) ([]run.RunEnd, error) {
	var ends []run.RunEnd
	err := s.session(ctx).Model(&models.TestProtocol{}).Select("test_run_id", "ended_at").
		Where("test_run_id IN ? AND ended_at IS NOT NULL", runIDs).Scan(&ends).Error
	return ends, storeerr.Map(err, "run ends")
}

// preloadRun loads what the run page shows of a run.
func preloadRun(db *gorm.DB) *gorm.DB {
	return db.Preload("DeviceStatus").Preload("DeviceStatus.Device").
		Preload("Protocols").Preload("Protocols.Device").Preload("Protocols.Entries").
		Preload("Protocols.Performance").
		Preload("Protocols.Keys", func(db *gorm.DB) *gorm.DB { return db.Order("key") }).
		Preload("Log").Preload("AppBinary").Preload("Test").Preload("Test.TestConfig")
}

// LastRun loads the newest run of a test in detail.
func (s *Store) LastRun(ctx context.Context, testID uint) (*models.TestRun, error) {
	var found models.TestRun
	err := preloadRun(s.session(ctx)).Where("test_id = ?", testID).Order("id desc").First(&found).Error
	if err != nil {
		return nil, lookup(err, "last run", apperr.NotFound("%s", gorm.ErrRecordNotFound.Error()))
	}
	return &found, nil
}

// Run loads one run in detail, with its recordings.
func (s *Store) Run(ctx context.Context, runID uint) (*models.TestRun, error) {
	var found models.TestRun
	err := preloadRun(s.session(ctx)).
		Preload("Recordings", func(db *gorm.DB) *gorm.DB { return db.Order("started_at") }).
		First(&found, runID).Error
	if err != nil {
		return nil, lookup(err, "run", apperr.NotFound("run %d not found", runID))
	}
	return &found, nil
}

// Neighbours finds the runs of the test before and after a run, by their
// creation time.
func (s *Store) Neighbours(ctx context.Context, testID, runID uint) (run.Neighbours, error) {
	runs := func() *gorm.DB { return s.session(ctx).Table("test_runs").Limit(1).Select("id") }
	prev, err := scanID(runs().Where("test_id = ? and id < ?", testID, runID).Order("created_at desc"))
	if err != nil {
		return run.Neighbours{}, err
	}
	next, err := scanID(runs().Where("test_id = ? and id > ?", testID, runID).Order("created_at asc"))
	return run.Neighbours{Prev: prev, Next: next}, err
}

// scanID reads the id the query selects; 0 when it finds none.
func scanID(query *gorm.DB) (uint, error) {
	var id uint
	err := query.Scan(&id).Error
	return id, storeerr.Map(err, "neighbouring run")
}

// StartupTimes lists the startup times of the device of a status in the
// earlier runs of the app.
func (s *Store) StartupTimes(ctx context.Context, q run.StartupQuery) ([]models.TestRunDeviceStatus, error) {
	var earlier []models.TestRunDeviceStatus
	err := s.session(ctx).Where("device_id = ? and test_run_id in (select tr.id from test_runs tr where tr.id < ? and tr.test_id in "+
		appTests+" order by tr.id desc limit ?)", q.Status.DeviceID, q.Status.TestRunID, q.AppID, q.Limit).Find(&earlier).Error
	return earlier, storeerr.Map(err, "startup times")
}
