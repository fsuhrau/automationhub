package models

import "gorm.io/gorm"

type Company struct {
	gorm.Model
	Token string
	Name  string
	Users []*User `gorm:"many2many:user_companies;"`
}
