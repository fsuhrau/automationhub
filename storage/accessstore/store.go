// Package accessstore is the gorm repository of users, access tokens and
// nodes. It implements service/access.Repo and maps gorm errors to apperr.
package accessstore

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

// Store reads and writes users, tokens and nodes.
type Store struct {
	db *gorm.DB
}

// New returns a Store on the database.
func New(db *gorm.DB) *Store { return &Store{db: db} }

// deleted is the result of a delete about what ("token 3"): a delete that
// removed nothing is a NotFound error.
func deleted(result *gorm.DB, what string) error {
	if result.Error != nil {
		return storeerr.Map(result.Error, what)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound("%s not found", what)
	}
	return nil
}

// CreateToken stores the token.
func (s *Store) CreateToken(ctx context.Context, token *models.AccessToken) error {
	return storeerr.Map(s.db.WithContext(ctx).Create(token).Error, "token "+token.Name)
}

// ProjectTokens lists the access tokens of the project.
func (s *Store) ProjectTokens(ctx context.Context, projectID uint) ([]models.AccessToken, error) {
	var tokens []models.AccessToken
	err := s.db.WithContext(ctx).Find(&tokens, "project_id = ?", projectID).Error
	return tokens, storeerr.Map(err, "tokens")
}

// DeleteProjectToken deletes a token of the project and nothing else.
func (s *Store) DeleteProjectToken(ctx context.Context, projectID, tokenID uint) error {
	result := s.db.WithContext(ctx).Where("id = ? AND project_id = ?", tokenID, projectID).Delete(&models.AccessToken{})
	return deleted(result, fmt.Sprintf("token %d", tokenID))
}

// PersonalTokens lists the tokens of the user, newest first.
func (s *Store) PersonalTokens(ctx context.Context, userID uint) ([]models.AccessToken, error) {
	var tokens []models.AccessToken
	err := s.db.WithContext(ctx).Order("id desc").Find(&tokens, "user_id = ?", userID).Error
	return tokens, storeerr.Map(err, "personal tokens")
}

// DeletePersonalToken deletes a token of the user and nothing else.
func (s *Store) DeletePersonalToken(ctx context.Context, userID, tokenID uint) error {
	result := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", tokenID, userID).Delete(&models.AccessToken{})
	return deleted(result, "token")
}
