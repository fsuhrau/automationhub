package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name string
	Auth []UserAuth
	Companies []*Company `gorm:"many2many:user_companies;"`
}