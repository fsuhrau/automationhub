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

// FindUser loads the stored user of a session. Sessions only carry name,
// login and email: the name is unique, the login is the provider user id
// for OAuth2 logins.
func (s *Store) FindUser(ctx context.Context, who authentication.User) (*models.User, error) {
	db := s.db.WithContext(ctx)
	if who.Name != "" {
		var user models.User
		if err := db.First(&user, "name = ?", who.Name).Error; err == nil {
			return &user, nil
		}
	}
	for _, q := range []struct{ query, value string }{
		{"provider_user_id = ?", who.Login},
		{"email = ?", who.Email},
	} {
		if user := userByAuth(db, q.query, q.value); user != nil {
			return user, nil
		}
	}
	return nil, storeerr.Map(gorm.ErrRecordNotFound, "user")
}

// userByAuth finds the user that owns the login method matching the query,
// nil if there is none.
func userByAuth(db *gorm.DB, query, value string) *models.User {
	if value == "" {
		return nil
	}
	var auth models.UserAuth
	if err := db.Preload("User").First(&auth, query, value).Error; err != nil {
		return nil
	}
	return auth.User
}
