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

// ensure creates the account of profile when the provider's user is new.
func (u users) ensure(p Profile) error {
	found, err := u.exists(p)
	if err != nil || found {
		return err
	}
	user := models.User{
		Name: p.Name,
		Auth: []models.UserAuth{{ProviderUserID: p.ProviderUserID, Email: p.Email, Provider: u.provider}},
	}
	createErr := u.db.Create(&user).Error
	if createErr == nil {
		return nil
	}
	// a parallel login of the same user may have created it first
	if found, err := u.exists(p); err == nil && found {
		return nil
	}
	return fmt.Errorf("create user: %w", createErr)
}

// exists tells whether the provider's user has an account.
func (u users) exists(p Profile) (bool, error) {
	var auth models.UserAuth
	err := u.db.First(&auth, "provider = ? and provider_user_id = ?", u.provider, p.ProviderUserID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("find user: %w", err)
	}
	return true, nil
}
