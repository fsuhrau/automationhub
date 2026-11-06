package api

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester"
	"github.com/fsuhrau/automationhub/tester/scenario"
	"github.com/fsuhrau/automationhub/tester/unity"
	"gorm.io/gorm"
)

// ErrNotFound marks lookups that failed because the entity doesn't exist or
// isn't part of the project.
var ErrNotFound = errors.New("not found")

// ErrBundleExists is returned when an identical app bundle was uploaded before.
var ErrBundleExists = errors.New("app bundle exists already")

func notFound(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrNotFound, fmt.Sprintf(format, args...))
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
	return run, nil
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
// creates the bundle entry. filePath must be inside apps.AppStoragePath.
func (s *Service) StoreBundle(application *models.App, filePath string) (*models.AppBinary, error) {
	analyser := app.NewAnalyser(filePath)
	if err := analyser.AnalyseFile(); err != nil {
		_ = os.Remove(filePath)
		return nil, fmt.Errorf("unable to analyse app: %w", err)
	}
	params := analyser.GetParameter()

	var existing models.AppBinary
	if err := s.db.First(&existing, "hash = ? AND app_id = ?", params.App.Hash, application.ID).Error; err == nil {
		_ = os.Remove(filePath)
		return &existing, ErrBundleExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	filename := filepath.Base(filePath)
	ext := filepath.Ext(filename)
	params.App.AppPath = fmt.Sprintf("%s_%s%s", strings.TrimSuffix(filename, ext), params.App.Hash, ext)
	if err := os.Rename(filePath, filepath.Join(apps.AppStoragePath, params.App.AppPath)); err != nil {
		return nil, fmt.Errorf("unable to store file: %w", err)
	}

	var launchActivity, executable string
	if params.App.Android != nil {
		launchActivity = params.App.Android.LaunchActivity
	}
	if params.App.Executable != nil {
		executable = params.App.Executable.Executable
	}

	bundle := models.AppBinary{
		AppPath:    params.App.AppPath,
		Hash:       params.App.Hash,
		Name:       params.Name,
		Identifier: params.Identifier,
		AppID:      application.ID,
		Version:    params.Version,
		Android:    models.Android{LaunchActivity: launchActivity},
		Executable: models.Executable{Executable: executable},
		Platform:   params.Platform,
		Additional: params.App.Additional,
		Size:       params.App.Size,
	}
	if err := s.db.Create(&bundle).Error; err != nil {
		return nil, fmt.Errorf("unable to create app bundle: %w", err)
	}
	return &bundle, nil
}
