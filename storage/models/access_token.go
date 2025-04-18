package models

import (
	"gorm.io/gorm"
	"time"
)

type AccessToken struct {
	gorm.Model
	ProjectID *uint
	NodeID    *uint
	Name      string
	Token     string
	ExpiresAt *time.Time
}
