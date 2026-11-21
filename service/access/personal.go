package access

import (
	"context"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/service/internal/validate"
	"github.com/fsuhrau/automationhub/storage/models"
)

const noUserMessage = "no signed in user; personal access tokens need a login (configure authentication on the master)"

// PersonalToken is a created personal token with its value, which is only
// handed out once: the store keeps a hash.
type PersonalToken struct {
	models.AccessToken
	Plain string
}

// Me resolves the stored user of a session. Without a signed in user, or
// for one the hub doesn't know, it is a NotFound error.
func (s *Service) Me(ctx context.Context, who authentication.User) (*models.User, error) {
	user, err := s.repo.FindUser(ctx, who)
	if apperr.KindOf(err) == apperr.KindNotFound {
		return nil, apperr.NotFound("%s", noUserMessage)
	}
	return user, err
}

// PersonalTokens lists the personal tokens of the user, newest first.
func (s *Service) PersonalTokens(ctx context.Context, who authentication.User) ([]models.AccessToken, error) {
	user, err := s.Me(ctx, who)
	if err != nil {
		return nil, err
	}
	return s.repo.PersonalTokens(ctx, user.ID)
}

// CreatePersonalToken creates a personal token for the user. The expiry,
// if there is one, has to be in the future.
func (s *Service) CreatePersonalToken(ctx context.Context, who authentication.User, in NewToken) (*PersonalToken, error) {
	user, err := s.Me(ctx, who)
	if err != nil {
		return nil, err
	}
	name, err := validate.TrimRequired(in.Name, "name is required")
	if err != nil {
		return nil, err
	}
	if in.ExpiresAt != nil && in.ExpiresAt.Before(s.now()) {
		return nil, apperr.Invalid("expiry is in the past")
	}
	plain, stored, hint, err := s.newPersonal()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, err, "create personal token")
	}
	created := models.AccessToken{UserID: &user.ID, Name: name, Token: stored, Hint: hint, ExpiresAt: in.ExpiresAt}
	if err := s.repo.CreateToken(ctx, &created); err != nil {
		return nil, err
	}
	return &PersonalToken{AccessToken: created, Plain: plain}, nil
}

// DeletePersonalToken revokes a token of the user; other users' tokens are
// not found.
func (s *Service) DeletePersonalToken(ctx context.Context, who authentication.User, tokenID uint) error {
	user, err := s.Me(ctx, who)
	if err != nil {
		return err
	}
	return s.repo.DeletePersonalToken(ctx, user.ID, tokenID)
}
