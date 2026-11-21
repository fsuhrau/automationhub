package oauthflow

import (
	"errors"
	"fmt"

	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

// users keeps the accounts of one provider in the database.
type users struct {
	db       *gorm.DB
	provider string
}

// ensure returns the account of profile and creates it when the provider's
// user is new. Accounts are matched by the provider's user id only.
func (u users) ensure(p Profile) (*models.User, error) {
	found, err := u.find(p)
	if err != nil || found != nil {
		return found, err
	}
	name, err := u.freeName(p)
	if err != nil {
		return nil, err
	}
	user := models.User{
		Name: name,
		Auth: []models.UserAuth{{ProviderUserID: p.ProviderUserID, Email: p.Email, Provider: u.provider}},
	}
	createErr := u.db.Create(&user).Error
	if createErr == nil {
		return &user, nil
	}
	// a parallel login of the same user may have created it first
	if found, err := u.find(p); err == nil && found != nil {
		return found, nil
	}
	return nil, fmt.Errorf("create user: %w", createErr)
}

// find returns the account of the provider's user, nil if there is none.
func (u users) find(p Profile) (*models.User, error) {
	var auth models.UserAuth
	err := u.db.Preload("User").First(&auth, "provider = ? and provider_user_id = ?", u.provider, p.ProviderUserID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return auth.User, nil
}
