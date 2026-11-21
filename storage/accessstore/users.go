package accessstore

import (
	"context"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

// Users lists all users.
func (s *Store) Users(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := s.db.WithContext(ctx).Find(&users).Error
	return users, storeerr.Map(err, "users")
}

// FindUser loads the stored user of a session by its hub user id. Names,
// logins and emails are labels the user may choose, so they never select an
// account; a session without an id finds nobody.
func (s *Store) FindUser(ctx context.Context, who authentication.User) (*models.User, error) {
	if who.ID == 0 {
		return nil, storeerr.Map(gorm.ErrRecordNotFound, "user")
	}
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, who.ID).Error; err != nil {
		return nil, storeerr.Map(err, "user")
	}
	return &user, nil
}
