package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name      string `gorm:"uniqueIndex;not null"`
	Role      string
	Auth      []UserAuth
	Companies []*Company `gorm:"many2many:user_companies;"`
	Projects  []*Project `gorm:"many2many:user_projects;"`
}
