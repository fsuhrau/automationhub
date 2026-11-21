// Package runnerstore is the gorm repository of the test runners: it creates
// the run, appends to its log and reads the configured unity tests. It
// implements tester/base.Store.
package runnerstore

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

// Store is the gorm implementation of the runner repository.
type Store struct {
	db *gorm.DB
}

// New builds the store on the database.
func New(db *gorm.DB) *Store { return &Store{db: db} }

// CreateRun inserts the run; its ID is set afterwards.
func (s *Store) CreateRun(ctx context.Context, run *models.TestRun) error {
	if err := s.db.WithContext(ctx).Create(run).Error; err != nil {
		return fmt.Errorf("create test run: %w", err)
	}
	return nil
}

// AddLog appends the entry to the log of its run.
func (s *Store) AddLog(ctx context.Context, entry *models.TestRunLogEntry) error {
	if err := s.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("add run log entry: %w", err)
	}
	return nil
}

// UnityTests returns the tests that are configured for the unity config.
func (s *Store) UnityTests(ctx context.Context, unityConfigID uint) ([]models.UnityTestFunction, error) {
	var tests []models.UnityTestFunction
	err := s.db.WithContext(ctx).Where("test_config_unity_id = ?", unityConfigID).Find(&tests).Error
	if err != nil {
		return nil, fmt.Errorf("load unity tests: %w", err)
	}
	return tests, nil
}
