// Package project is the use cases of the projects: listing, creating,
// renaming and deleting them.
package project

import (
	"context"

	"github.com/fsuhrau/automationhub/service/internal/validate"
	"github.com/fsuhrau/automationhub/storage/models"
)

// defaultIdentifier is the project the hub starts with; it is the only one
// whose identifier can be changed, since the clients know it.
const defaultIdentifier = "default_project"

// Repo is what the service needs from the storage. Lookups that find
// nothing return an apperr.NotFound error.
type Repo interface {
	// List returns all projects with their apps and the apps' parameters.
	List(ctx context.Context) ([]models.Project, error)
	// CreateUnique stores the project under the identifier base, or under
	// base-2, base-3 and so on when that is taken. Deleted projects keep
	// their identifier, so old links don't lead to a new project.
	CreateUnique(ctx context.Context, project *models.Project, base string) error
	// ByIdentifier loads the project with the identifier.
	ByIdentifier(ctx context.Context, identifier string) (*models.Project, error)
	// Save stores the changes of a loaded project.
	Save(ctx context.Context, project *models.Project) error
	// Delete deletes the project with the id.
	Delete(ctx context.Context, id uint) error
}

// Deps are the dependencies of the Service.
type Deps struct {
	Repo Repo
}

// Service implements the project use cases.
type Service struct {
	repo Repo
}

// New returns a Service on the repo.
func New(deps Deps) *Service { return &Service{repo: deps.Repo} }

// Change is what a rename asks for.
type Change struct {
	Name string
	// Identifier is only applied to the default project.
	Identifier string
}

// List returns all projects.
func (s *Service) List(ctx context.Context) ([]models.Project, error) {
	return s.repo.List(ctx)
}

// Create creates a project from its name. The identifier is derived from
// the name and made unique.
func (s *Service) Create(ctx context.Context, name string) (*models.Project, error) {
	name, err := validate.TrimRequired(name, "the project needs a name")
	if err != nil {
		return nil, err
	}
	project := &models.Project{Name: name}
	if err := s.repo.CreateUnique(ctx, project, slug(name)); err != nil {
		return nil, err
	}
	return project, nil
}

// Rename renames the project with the identifier. Only the default project
// takes a new identifier too.
func (s *Service) Rename(ctx context.Context, identifier string, change Change) (*models.Project, error) {
	name, err := validate.TrimRequired(change.Name, "the project needs a name")
	if err != nil {
		return nil, err
	}
	project, err := s.repo.ByIdentifier(ctx, identifier)
	if err != nil {
		return nil, err
	}
	project.Name = name
	if project.Identifier == defaultIdentifier && change.Identifier != "" {
		project.Identifier = change.Identifier
	}
	if err := s.repo.Save(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

// Delete deletes the project with the id.
func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
