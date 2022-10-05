package models

import (
	"gorm.io/gorm"
)

type App struct {
	gorm.Model
	ProjectID        uint
	Project          *Project
	Name             string
	Identifier       string
	Platform         PlatformType
	DefaultParameter string
}
