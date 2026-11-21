// Package appstore is the gorm repository of apps, their parameters and
// their bundles. It implements service/app.Repo and maps gorm errors to
// apperr.
package appstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/fsuhrau/automationhub/apperr"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

// Store is the gorm implementation of the app repository.
type Store struct {
	db *gorm.DB
}

// New builds the store on the database.
func New(db *gorm.DB) *Store { return &Store{db: db} }

var _ appsvc.Repo = (*Store)(nil)

func (s *Store) tx(ctx context.Context) *gorm.DB { return s.db.WithContext(ctx) }

// lookup maps the error of a single-record query: a missing record is
// "<what> not found", without the storage detail the message would otherwise
// carry to the person who sent the request.
func lookup(err error, what string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.NotFound("%s not found", what)
	}
	return storeerr.Map(err, what)
}

// inScope limits a query to the rows of the app column that belong to the
// app of the scope and to the project that owns it.
func inScope(db *gorm.DB, column string, sc appsvc.Scope) *gorm.DB {
	return db.Where(column+" IN (SELECT id FROM apps WHERE id = ? AND project_id = ? AND deleted_at IS NULL)", sc.AppID, sc.ProjectID)
}

// Apps lists the apps of the project with their parameters.
func (s *Store) Apps(ctx context.Context, projectID uint) ([]models.App, error) {
	var found []models.App
	err := s.tx(ctx).Where("project_id = ?", projectID).Preload("Parameter").Find(&found).Error
	return found, storeerr.Map(err, "apps")
}

// App loads one app of the project with its parameters.
func (s *Store) App(ctx context.Context, sc appsvc.Scope) (*models.App, error) {
	var found models.App
	err := s.tx(ctx).Preload("Parameter").First(&found, "project_id = ? AND id = ?", sc.ProjectID, sc.AppID).Error
	return &found, lookup(err, fmt.Sprintf("app %d", sc.AppID))
}

// CreateApp inserts the app.
func (s *Store) CreateApp(ctx context.Context, app *models.App) error {
	return storeerr.Map(s.tx(ctx).Create(app).Error, "app")
}

// SaveApp writes the app; its parameters are left alone.
func (s *Store) SaveApp(ctx context.Context, app *models.App) error {
	err := s.tx(ctx).Omit("Parameter").Save(app).Error
	return storeerr.Map(err, fmt.Sprintf("app %d", app.ID))
}
