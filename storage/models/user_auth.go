package models

import "gorm.io/gorm"

type UserAuth struct {
	gorm.Model
	UserID         uint `gorm:"uniqueIndex:idx_userid_provider;not null"`
	User           *User
	Email          string `gorm:"uniqueIndex:idx_email_provider;not null"`
	Provider       string `gorm:"uniqueIndex:idx_email_provider;uniqueIndex:idx_userid_provider;not null"`
	ProviderUserID string
	Token          string
	Username       string
	PasswordHash   string
}
