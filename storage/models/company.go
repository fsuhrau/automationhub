package models

import "gorm.io/gorm"

type Company struct {
	gorm.Model
	Name string
	Users []*User `gorm:"many2many:user_companies;"`
}