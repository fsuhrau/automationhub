package runstore

import (
	"context"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/devicestore"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
)

// StartTest loads a test of the project with its app, config, devices and
// Unity settings.
func (s *Store) StartTest(ctx context.Context, projectID, testID uint) (*models.Test, error) {
	var test models.Test
	err := s.session(ctx).Preload("App").Preload("TestConfig").Preload("TestConfig.Devices").
		Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").
		Joins("JOIN apps ON apps.id = tests.app_id").
		Where("apps.project_id = ?", projectID).
		First(&test, "tests.id = ?", testID).Error
	if err != nil {
		return nil, lookup(err, "test", apperr.NotFound("not found: test %d", testID))
	}
	return &test, nil
}

// Bundle loads a bundle of the app, with the app.
func (s *Store) Bundle(ctx context.Context, appID, bundleID uint) (*models.AppBinary, error) {
	var binary models.AppBinary
	err := s.session(ctx).Preload("App").First(&binary, "id = ? AND app_id = ?", bundleID, appID).Error
	if err != nil {
		return nil, lookup(err, "app bundle", apperr.NotFound("not found: app bundle %d for app %d", bundleID, appID))
	}
	return &binary, nil
}

// Devices lists the devices of the query with their node, by id.
func (s *Store) Devices(ctx context.Context, q run.DeviceQuery) ([]models.Device, error) {
	query := devicestore.VisibleTo(s.session(ctx), q.Project)
	if !q.All {
		query = query.Where("id IN ?", q.IDs)
	}
	var devices []models.Device
	err := query.Preload("Node").Order("id").Find(&devices).Error
	return devices, storeerr.Map(err, "devices")
}

// Device loads a device the project can see, with its node.
func (s *Store) Device(ctx context.Context, project *models.Project, id uint) (*models.Device, error) {
	var d models.Device
	err := devicestore.VisibleTo(s.session(ctx), project).Preload("Node").First(&d, "id = ?", id).Error
	if err != nil {
		return nil, lookup(err, "device", apperr.NotFound("not found: device %d", id))
	}
	return &d, nil
}

// RunInProject tells that the run belongs to the project.
func (s *Store) RunInProject(ctx context.Context, projectID, runID uint) error {
	query := s.session(ctx).Model(&models.TestRun{}).
		Joins("JOIN tests ON tests.id = test_runs.test_id").
		Joins("JOIN apps ON apps.id = tests.app_id").
		Where("apps.project_id = ? AND test_runs.id = ?", projectID, runID)
	return require(query, "test run", apperr.NotFound("not found: test run %d", runID))
}
