package models

import "gorm.io/gorm"

type Test struct {
	gorm.Model
	AppID uint
	App App
	Status string
}
