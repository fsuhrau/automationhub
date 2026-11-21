// Package projectstore is the gorm repository of the projects. It
// implements service/project.Repo and maps gorm errors to apperr.
package projectstore

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

// Store reads and writes projects.
type Store struct {
	db *gorm.DB
}

// New returns a Store on the database.
func New(db *gorm.DB) *Store { return &Store{db: db} }

// List returns all projects with their apps and the apps' parameters.
func (s *Store) List(ctx context.Context) ([]models.Project, error) {
	var projects []models.Project
	err := s.db.WithContext(ctx).Preload("Apps").Preload("Apps.Parameter").Find(&projects).Error
	return projects, storeerr.Map(err, "projects")
}

// CreateUnique stores the project under the identifier base, or under
// base-2, base-3 and so on when that is taken, also by a deleted project.
func (s *Store) CreateUnique(ctx context.Context, project *models.Project, base string) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		project.Identifier = base
		for i := 2; ; i++ {
			taken, err := identifierTaken(tx, project.Identifier)
			if err != nil {
				return err
			}
			if !taken {
				return tx.Create(project).Error
			}
			project.Identifier = fmt.Sprintf("%s-%d", base, i)
		}
	})
	return storeerr.Map(err, "project "+base)
}

// identifierTaken counts deleted projects too: they keep their identifier.
func identifierTaken(tx *gorm.DB, identifier string) (bool, error) {
	var taken int64
	err := tx.Unscoped().Model(&models.Project{}).Where("identifier = ?", identifier).Count(&taken).Error
	return taken > 0, err
}

// ByIdentifier loads the project with the identifier.
func (s *Store) ByIdentifier(ctx context.Context, identifier string) (*models.Project, error) {
	var project models.Project
	err := s.db.WithContext(ctx).First(&project, "identifier = ?", identifier).Error
	if err != nil {
		return nil, storeerr.Map(err, "project "+identifier)
	}
	return &project, nil
}

// Save stores the changes of a loaded project.
func (s *Store) Save(ctx context.Context, project *models.Project) error {
	err := s.db.WithContext(ctx).Save(project).Error
	return storeerr.Map(err, "project "+project.Identifier)
}

// Delete deletes the project with the id.
func (s *Store) Delete(ctx context.Context, id uint) error {
	what := fmt.Sprintf("project %d", id)
	var project models.Project
	if err := s.db.WithContext(ctx).First(&project, id).Error; err != nil {
		return storeerr.Map(err, what)
	}
	return storeerr.Map(s.db.WithContext(ctx).Delete(&project).Error, what)
}
