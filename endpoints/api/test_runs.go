package api

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/fsuhrau/automationhub/events"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/tester/scenario"
	"github.com/fsuhrau/automationhub/tester/unity"
	"gorm.io/gorm"
)

// ErrNotFound marks lookups that failed because the entity doesn't exist or
// isn't part of the project.
var ErrNotFound = errors.New("not found")

// ErrBundleExists is returned when an identical app bundle was uploaded before.
var ErrBundleExists = errors.New("app bundle exists already")

// ErrNoDevices is returned when none of a test's devices can run it.
var ErrNoDevices = errors.New("no device available to run the test")

func notFound(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrNotFound, fmt.Sprintf(format, args...))
}

// invalidRequest is a request the master refuses; its message is meant for
// the person who sent it
type invalidRequest struct{ message string }

func (e invalidRequest) Error() string { return e.message }

func badRequest(format string, args ...any) error {
	return invalidRequest{fmt.Sprintf(format, args...)}
}

// ProjectDevices lists the devices visible to the project.
func (s *Service) ProjectDevices(project *models.Project) ([]models.Device, error) {
	var devices []models.Device
	err := s.projectDevices(project).Preload("Node").Order("id").Find(&devices).Error
	return devices, err
}

// FindDevice returns one device of the project.
func (s *Service) FindDevice(project *models.Project, id uint) (*models.Device, error) {
	var d models.Device
	err := s.projectDevices(project).Preload("Node").First(&d, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, notFound("device %d", id)
	}
	return &d, err
}

// StartTestRun starts the test with the given bundle (or start URL for web
// and editor apps) on the devices of the test configuration.
func (s *Service) StartTestRun(project *models.Project, testID, binaryID uint, startURL, params string) (*models.TestRun, error) {
	var test models.Test
	err := s.db.Preload("App").Preload("TestConfig").Preload("TestConfig.Devices").
		Joins("JOIN apps ON apps.id = tests.app_id").
		Where("apps.project_id = ?", project.ID).
		First(&test, "tests.id = ?", testID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, notFound("test %d", testID)
	}
	if err != nil {
		return nil, err
	}

	var binary *models.AppBinary
	if test.App.Platform != models.PlatformTypeEditor && test.App.Platform != models.PlatformTypeWeb {
		binary = &models.AppBinary{}
		err := s.db.Preload("App").First(binary, "id = ? AND app_id = ?", binaryID, test.AppID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFound("app bundle %d for app %d", binaryID, test.AppID)
		}
		if err != nil {
			return nil, err
		}
		if binary.Archived() {
			return nil, badRequest("app bundle %d is archived; upload its file again to run tests with it", binaryID)
		}
	}

	var devices []models.Device
	query := s.projectDevices(project)
	if !test.TestConfig.AllDevices {
		query = query.Where("id IN ?", test.TestConfig.GetDeviceIds())
	}
	if err := query.Find(&devices).Error; err != nil {
		return nil, err
	}
	for i := range devices {
		devices[i].Dev, _ = s.devicesManager.GetDevice(devices[i].DeviceIdentifier)
	}
	if err := checkDevices(devices); err != nil {
		return nil, err
	}

	var testRunner tester.Interface
	switch test.TestConfig.Type {
	case models.TestTypeUnity:
		testRunner = unity.New(s.db, s.nodeUrl, s.devicesManager, s, project.Identifier, test.AppID)
		if err := s.db.Preload("UnityTestFunctions").Where("test_config_id = ?", test.TestConfig.ID).First(&test.TestConfig.Unity).Error; err != nil {
			return nil, err
		}
	case models.TestTypeScenario:
		testRunner = scenario.New(s.db, s.nodeUrl, s.devicesManager, s, project.Identifier, test.AppID)
	default:
		return nil, fmt.Errorf("invalid test config type")
	}

	if err := testRunner.Initialize(test, extractParams(params)); err != nil {
		return nil, err
	}

	run, err := testRunner.Run(devices, binary, startURL)
	if err != nil {
		return nil, err
	}

	s.runnersMutex.Lock()
	s.runners[runKey(run.ID)] = testRunner
	s.runnersMutex.Unlock()
	s.runTracker.Started(run.ID)
	events.TestRunStarted.Trigger(events.TestRunStartedPayload{
		ProjectID: project.Identifier,
		AppID:     test.AppID,
		TestID:    test.ID,
		TestRunID: run.ID,
	})
	return run, nil
}

// checkDevices refuses a run when none of its devices can run tests, so the
// caller (web, CLI, MCP, CI) fails right away instead of getting an empty
// run. Unavailable devices next to available ones are skipped by the run.
func checkDevices(devices []models.Device) error {
	if len(devices) == 0 {
		return fmt.Errorf("%w: the test has no devices", ErrNoDevices)
	}
	reasons := make([]string, 0, len(devices))
	for _, d := range devices {
		reason := base.Unavailable(d)
		if reason == "" {
			return nil
		}
		name := d.Alias
		if name == "" {
			name = d.Name
		}
		if name == "" {
			name = d.DeviceIdentifier
		}
		reasons = append(reasons, fmt.Sprintf("%s is %s", name, reason))
	}
	return fmt.Errorf("%w: %s", ErrNoDevices, strings.Join(reasons, "; "))
}

func runKey(runID uint) string {
	return strconv.FormatUint(uint64(runID), 10)
}

// CancelTestRun cancels a run started by this master.
func (s *Service) CancelTestRun(project *models.Project, runID uint) error {
	if _, err := s.projectTestRun(project, runID); err != nil {
		return err
	}
	s.runnersMutex.Lock()
	testRunner, exists := s.runners[runKey(runID)]
	s.runnersMutex.Unlock()
	if !exists {
		return notFound("no active runner for test run %d", runID)
	}
	if err := testRunner.Cancel(runKey(runID)); err != nil {
		return err
	}
	s.runnersMutex.Lock()
	delete(s.runners, runKey(runID))
	s.runnersMutex.Unlock()
	return nil
}

func (s *Service) onRunFinished(runID uint) {
	s.runnersMutex.Lock()
	delete(s.runners, runKey(runID))
	s.runnersMutex.Unlock()
}

// projectTestRun loads a test run if it belongs to the project.
func (s *Service) projectTestRun(project *models.Project, runID uint) (*models.TestRun, error) {
	var run models.TestRun
	err := s.db.Joins("JOIN tests ON tests.id = test_runs.test_id").
		Joins("JOIN apps ON apps.id = tests.app_id").
		Where("apps.project_id = ?", project.ID).
		First(&run, "test_runs.id = ?", runID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, notFound("test run %d", runID)
	}
	return &run, err
}

// StoreBundle analyses an uploaded app file, moves it into the app storage and
// creates the bundle entry. filePath must be inside apps.AppStoragePath. An
// identical bundle comes back with ErrBundleExists.
func (s *Service) StoreBundle(application *models.App, filePath string) (*models.AppBinary, error) {
	scope := appsvc.Scope{ProjectID: application.ProjectID, AppID: application.ID}
	bundle, err := s.appService().StoreBundle(context.Background(), scope, filePath)
	if errors.Is(err, appsvc.ErrBundleExists) {
		return bundle, ErrBundleExists
	}
	return bundle, err
}
