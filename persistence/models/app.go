package models

import (
	"gorm.io/gorm"
)

type App struct {
	gorm.Model
	Name string
	Filename string
	Platform string
	Version string
	AppID string
}