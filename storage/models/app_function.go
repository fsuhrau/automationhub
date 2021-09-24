package models

import (
	"gorm.io/gorm"
)

type AppFunction struct {
	gorm.Model
	AppID    string
	Assembly string
	Class    string
	Method   string
}
