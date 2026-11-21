package test

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Create creates a test with the edit page's rules; only Unity tests can be
// created here.
func (s *Service) Create(ctx context.Context, sc Scope, in NewTest) (*models.Test, error) {
	if in.Type != models.TestTypeUnity {
		return nil, apperr.Invalid("only Unity tests can be created here")
	}
	if err := in.validate(true); err != nil {
		return nil, err
	}
	if err := s.requireApp(ctx, sc); err != nil {
		return nil, err
	}
	if err := s.checkDevices(ctx, sc, in.Devices); err != nil {
		return nil, err
	}
	test := models.Test{AppID: sc.AppID, Name: in.Name, TestConfig: config(in.Type, in.Edit, true)}
	created, err := s.repo.CreateTest(ctx, test)
	if err != nil {
		return nil, fmt.Errorf("create test: %w", err)
	}
	return &created, nil
}

// Update saves the test edit page in one transaction: the devices and the
// Unity test functions are replaced by the submitted ones. The test must
// belong to the app and the devices to the project.
func (s *Service) Update(ctx context.Context, ref TestRef, in Edit) (*models.Test, error) {
	current, err := s.repo.Test(ctx, ref)
	if err != nil {
		return nil, err
	}
	unity := current.TestConfig.Type == models.TestTypeUnity && current.TestConfig.Unity != nil
	if err := in.validate(unity); err != nil {
		return nil, err
	}
	if err := s.checkDevices(ctx, ref.Scope, in.Devices); err != nil {
		return nil, err
	}
	desired := models.Test{Name: in.Name, TestConfig: config(current.TestConfig.Type, in, unity)}
	if err := s.repo.UpdateTest(ctx, current, desired); err != nil {
		return nil, fmt.Errorf("update test %d: %w", ref.TestID, err)
	}
	return s.repo.Test(ctx, ref)
}

// requireApp makes sure the app belongs to the project.
func (s *Service) requireApp(ctx context.Context, sc Scope) error {
	ok, err := s.repo.AppExists(ctx, sc)
	if err != nil {
		return fmt.Errorf("look up app %d: %w", sc.AppID, err)
	}
	if !ok {
		return apperr.NotFound("app %d not found", sc.AppID)
	}
	return nil
}

// checkDevices makes sure every device exists and is visible to the project.
func (s *Service) checkDevices(ctx context.Context, sc Scope, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	visible, err := s.repo.CountDevices(ctx, sc.ProjectID, ids)
	if err != nil {
		return fmt.Errorf("check devices: %w", err)
	}
	if visible != len(ids) {
		return apperr.Invalid("some of the picked devices don't exist")
	}
	return nil
}

// config is the test config the edit asks for.
func config(typ models.TestType, in Edit, unity bool) models.TestConfig {
	cfg := models.TestConfig{
		Type: typ, AllDevices: in.AllDevices, ExecutionType: in.ExecutionType, RecordSession: in.RecordSession,
	}
	for _, d := range in.Devices {
		cfg.Devices = append(cfg.Devices, models.TestConfigDevice{DeviceID: d})
	}
	if unity {
		cfg.Unity = &models.TestConfigUnity{
			UnityTestCategoryType: in.UnityTestCategoryType, Categories: in.Categories, UnityTestFunctions: in.TestFunctions,
		}
	}
	return cfg
}
