package access

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/internal/validate"
	"github.com/fsuhrau/automationhub/storage/models"
)

// NewToken is what a token is created from.
type NewToken struct {
	Name      string
	ExpiresAt *time.Time
}

// CreateProjectToken creates an access token for the project. Its value is
// stored as it is; project tokens are shown in the settings.
func (s *Service) CreateProjectToken(ctx context.Context, projectID uint, in NewToken) (*models.AccessToken, error) {
	name, err := validate.TrimRequired(in.Name, "the token needs a name")
	if err != nil {
		return nil, err
	}
	value, err := s.newUUID()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, err, "create token value")
	}
	created := &models.AccessToken{ProjectID: &projectID, Token: value, ExpiresAt: in.ExpiresAt, Name: name}
	if err := s.repo.CreateToken(ctx, created); err != nil {
		return nil, err
	}
	return created, nil
}

// ProjectTokens lists the access tokens of the project.
func (s *Service) ProjectTokens(ctx context.Context, projectID uint) ([]models.AccessToken, error) {
	return s.repo.ProjectTokens(ctx, projectID)
}

// DeleteProjectToken revokes a token of the project. Tokens of other
// projects, of nodes and personal ones are not found.
func (s *Service) DeleteProjectToken(ctx context.Context, projectID, tokenID uint) error {
	return s.repo.DeleteProjectToken(ctx, projectID, tokenID)
}
