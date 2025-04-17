package models

import (
	"gorm.io/gorm"
)

type Node struct {
	gorm.Model
	Identifier string
	Name       string
	Status     int `db:"-"`
}
