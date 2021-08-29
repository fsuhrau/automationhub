package models

import (
	"gorm.io/gorm"
)

type App struct {
	gorm.Model
	AppID string
	Name string
	Filename string
	Platform string
	Version string
}