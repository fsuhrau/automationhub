package teststore

import (
	"context"

	"github.com/fsuhrau/automationhub/apperr"
	service "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

var _ service.Repo = (*Store)(nil)

// Store reads and writes the test definitions.
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

// preloadConfig loads the config of the tests: devices, Unity settings and
// the Unity test functions.
func preloadConfig(db *gorm.DB) *gorm.DB {
	return db.Preload("TestConfig").Preload("TestConfig.Devices").
		Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions")
}

// preloadDetail is preloadConfig and the device behind each config device.
func preloadDetail(db *gorm.DB) *gorm.DB {
	return preloadConfig(db).Preload("TestConfig.Devices.Device")
}

// inScope limits a query on tests to the tests of the app, and the app to
// the project: a test of another project is not in scope.
func inScope(sc service.Scope) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tests.app_id = ? AND tests.app_id IN (?)", sc.AppID,
			db.Session(&gorm.Session{NewDB: true}).Model(&models.App{}).Select("id").Where("project_id = ?", sc.ProjectID))
	}
}

// Tests lists the tests of the scope with their config, ordered by id.
func (s *Store) Tests(ctx context.Context, sc service.Scope) ([]models.Test, error) {
	var tests []models.Test
	err := preloadConfig(s.session(ctx).Scopes(inScope(sc))).Order("tests.id").Find(&tests).Error
	return tests, storeerr.Map(err, "tests")
}

// Test loads one test of the scope with its config in detail. A missing
// test is a NotFound error with the message "record not found", which the
// api has always answered with.
func (s *Store) Test(ctx context.Context, ref service.TestRef) (*models.Test, error) {
	var test models.Test
	err := preloadDetail(s.session(ctx).Scopes(inScope(ref.Scope))).First(&test, "tests.id = ?", ref.TestID).Error
	if err != nil {
		return nil, mapLookup(err, "test")
	}
	return &test, nil
}

// mapLookup is storeerr.Map except for a missing row, which keeps gorm's
// message "record not found" as the api's contract has it.
func mapLookup(err error, what string) error {
	if mapped := storeerr.Map(err, what); apperr.KindOf(mapped) != apperr.KindNotFound {
		return mapped
	}
	return apperr.NotFound("%s", gorm.ErrRecordNotFound.Error())
}

// AppExists tells whether the app belongs to the project.
func (s *Store) AppExists(ctx context.Context, sc service.Scope) (bool, error) {
	var n int64
	err := s.session(ctx).Model(&models.App{}).Where("id = ? AND project_id = ?", sc.AppID, sc.ProjectID).Count(&n).Error
	return n > 0, storeerr.Map(err, "app")
}

// CountDevices counts how many of the devices the project can see: those of
// its company and the ones that belong to no company.
func (s *Store) CountDevices(ctx context.Context, projectID uint, ids []uint) (int, error) {
	var n int64
	company := s.session(ctx).Session(&gorm.Session{NewDB: true}).Model(&models.Project{}).Select("company_id").Where("id = ?", projectID)
	err := s.session(ctx).Model(&models.Device{}).
		Where("id IN ? AND (company_id = 0 OR company_id IN (?))", ids, company).Count(&n).Error
	return int(n), storeerr.Map(err, "devices")
}
