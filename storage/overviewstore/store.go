// Package overviewstore is the gorm repository of the project overview and
// statistics. It implements service/overview.Repo and maps gorm errors to
// apperr.
package overviewstore

import (
	"context"

	"github.com/fsuhrau/automationhub/service/overview"
	"github.com/fsuhrau/automationhub/storage/devicestore"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

var _ overview.Repo = (*Store)(nil)

// Store reads what the overview and the statistics show.
type Store struct {
	db *gorm.DB
}

// New builds the store on the database.
func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) session(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}

// appsOf selects the ids of the apps of the project.
func (s *Store) appsOf(ctx context.Context, projectID uint) *gorm.DB {
	return s.session(ctx).Session(&gorm.Session{NewDB: true}).Model(&models.App{}).Select("id").Where("project_id = ?", projectID)
}

// Totals counts the project's apps and bundles and sums the bytes of the
// bundles that still have their file.
func (s *Store) Totals(ctx context.Context, projectID uint) (overview.Totals, error) {
	var totals overview.Totals
	err := s.session(ctx).Model(&models.App{}).Where("project_id = ?", projectID).Count(&totals.Apps).Error
	if err != nil {
		return totals, storeerr.Map(err, "apps")
	}
	var bundles struct {
		N    int64
		Size int64
	}
	err = s.session(ctx).Model(&models.AppBinary{}).
		Select("count(*) as n, coalesce(sum(case when archived_at is null then size else 0 end), 0) as size").
		Where("app_id IN (?)", s.appsOf(ctx, projectID)).Scan(&bundles).Error
	totals.Bundles, totals.BundleStorage = bundles.N, bundles.Size
	return totals, storeerr.Map(err, "bundles")
}

// Tests lists the tests of the project's apps, by id.
func (s *Store) Tests(ctx context.Context, projectID uint) ([]models.Test, error) {
	var tests []models.Test
	err := s.session(ctx).Where("app_id IN (?)", s.appsOf(ctx, projectID)).Order("id").Find(&tests).Error
	return tests, storeerr.Map(err, "tests")
}

// LastRuns lists the newest run of each test, newest first.
func (s *Store) LastRuns(ctx context.Context, testIDs []uint) ([]models.TestRun, error) {
	newest := s.session(ctx).Session(&gorm.Session{NewDB: true}).Model(&models.TestRun{}).
		Select("max(id)").Where("test_id IN ?", testIDs).Group("test_id")
	var found []models.TestRun
	err := s.session(ctx).Where("id IN (?)", newest).Order("id desc").Find(&found).Error
	return found, storeerr.Map(err, "last runs")
}

// LatestRuns lists the newest runs of the tests, newest first.
func (s *Store) LatestRuns(ctx context.Context, testIDs []uint, limit int) ([]models.TestRun, error) {
	var found []models.TestRun
	err := s.session(ctx).Where("test_id IN ?", testIDs).Order("id desc").Limit(limit).Find(&found).Error
	return found, storeerr.Map(err, "latest runs")
}

// Protocols lists the newest protocols of the project with their run, test
// and device.
func (s *Store) Protocols(ctx context.Context, q overview.ProtocolQuery) ([]models.TestProtocol, error) {
	runs := s.session(ctx).Session(&gorm.Session{NewDB: true}).Model(&models.TestRun{}).Select("test_runs.id").
		Joins("JOIN tests ON tests.id = test_runs.test_id").
		Where("tests.app_id IN (?)", s.appsOf(ctx, q.ProjectID))
	query := s.session(ctx).Preload("TestRun").Preload("TestRun.Test").Preload("Device").
		Where("test_run_id IN (?)", runs)
	if q.FailedOnly {
		query = query.Where("test_result = ?", models.TestResultFailed)
	}
	var found []models.TestProtocol
	err := query.Order("created_at desc").Limit(q.Limit).Find(&found).Error
	return found, storeerr.Map(err, "protocols")
}

// Devices lists the devices the project can see.
func (s *Store) Devices(ctx context.Context, project *models.Project) ([]models.Device, error) {
	var devices []models.Device
	err := devicestore.VisibleTo(s.session(ctx), project).Find(&devices).Error
	return devices, storeerr.Map(err, "devices")
}
