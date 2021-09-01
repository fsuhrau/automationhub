package models

import (
	"gorm.io/gorm"
)

type App struct {
	gorm.Model
	AppID          string
	Name           string
	Platform       string
	Version        string
	AppPath        string
	Identifier     string
	LaunchActivity string
	Additional     string
	Hash           [20]byte
}
