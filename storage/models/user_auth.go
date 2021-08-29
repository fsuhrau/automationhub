package models

import "gorm.io/gorm"

type UserAuth struct {
	gorm.Model
	UserID       uint
	Provider     string
	Token        string
	Username     string
	Email        string
	PasswordHash string
}
