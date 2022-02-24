package models

import (
	"gorm.io/gorm"
	"time"
)

type AccessToken struct {
	gorm.Model
	CompanyID uint
	Name      string
	Token     string
	ExpiresAt *time.Time
}
