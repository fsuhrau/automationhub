// Package devicestore is the gorm repository of devices. It implements
// service/device.Repo and maps gorm errors to apperr.
package devicestore

import (
	"context"
	"fmt"
	"strconv"

	svc "github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/teststore"
	"gorm.io/gorm"
)

var _ svc.Repo = (*Store)(nil)

// Store reads and writes devices.
type Store struct {
	db *gorm.DB
}

// New returns a Store on db.
func New(db *gorm.DB) *Store { return &Store{db: db} }

// VisibleTo scopes a query on devices to the project's company. Devices that
// are not assigned to a company (the default for auto-discovered devices)
// are visible to every project.
func VisibleTo(db *gorm.DB, project *models.Project) *gorm.DB {
	return db.Where("company_id IN ?", []uint{0, project.CompanyID})
}

// List returns the devices visible to the project with their node, limited
// to a platform when one is given.
func (s *Store) List(ctx context.Context, project *models.Project, platform *models.PlatformType) ([]models.Device, error) {
	query := VisibleTo(s.db.WithContext(ctx), project)
	if platform != nil {
		query = query.Where("platform_type = ?", *platform)
	}
	var devices []models.Device
	if err := query.Preload("Node").Find(&devices).Error; err != nil {
		return nil, mapErr(err, "devices")
	}
	return devices, nil
}

// Find returns the device with the id if the project can see it.
func (s *Store) Find(ctx context.Context, project *models.Project, id string) (*models.Device, error) {
	return s.find(s.db.WithContext(ctx), project, id)
}

// FindDetail is Find with the connection, device and custom parameters.
func (s *Store) FindDetail(ctx context.Context, project *models.Project, id string) (*models.Device, error) {
	db := s.db.WithContext(ctx).Preload("ConnectionParameter").Preload("DeviceParameter").Preload("CustomParameter")
	return s.find(db, project, id)
}

// find loads the visible device from db, which carries the context and the
// preloads.
func (s *Store) find(db *gorm.DB, project *models.Project, id string) (*models.Device, error) {
	var d models.Device
	if err := VisibleTo(db, project).First(&d, "id = ?", id).Error; err != nil {
		return nil, mapErr(err, "device "+id)
	}
	return &d, nil
}

// Delete removes the device.
func (s *Store) Delete(ctx context.Context, d *models.Device) error {
	return mapErr(s.db.WithContext(ctx).Delete(d).Error, "device "+strconv.FormatUint(uint64(d.ID), 10))
}

// Protocols returns the newest results of tests of the project on the
// device. A runner protocol that wraps a test has the same name and would
// show every test twice, so only the tests themselves count.
func (s *Store) Protocols(ctx context.Context, scope svc.ProtocolScope) ([]models.TestProtocol, error) {
	query := s.db.WithContext(ctx).
		Joins("JOIN test_runs ON test_runs.id = test_protocols.test_run_id").
		Joins("JOIN tests ON tests.id = test_runs.test_id").
		Joins("JOIN apps ON apps.id = tests.app_id").
		Where("test_protocols.device_id = ? AND apps.project_id = ?", scope.DeviceID, scope.ProjectID)
	var protocols []models.TestProtocol
	err := teststore.LeafProtocols(query).
		Preload("TestRun").Preload("TestRun.Test").
		Order("test_protocols.id desc").Limit(scope.Limit).
		Find(&protocols).Error
	return protocols, mapErr(err, fmt.Sprintf("protocols of device %d", scope.DeviceID))
}

// Bundle returns the app bundle with the id if its app belongs to the project.
func (s *Store) Bundle(ctx context.Context, project *models.Project, id uint) (*models.AppBinary, error) {
	var binary models.AppBinary
	err := s.db.WithContext(ctx).Preload("App").
		Joins("JOIN apps ON apps.id = app_binaries.app_id").
		Where("apps.project_id = ?", project.ID).
		First(&binary, "app_binaries.id = ?", id).Error
	if err != nil {
		return nil, mapErr(err, fmt.Sprintf("app bundle %d", id))
	}
	return &binary, nil
}
