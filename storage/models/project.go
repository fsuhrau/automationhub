package models

import "gorm.io/gorm"

type Project struct {
	gorm.Model
	Identifier   string
	Name         string
	CompanyID    uint
	Company      *Company
	AccessTokens []*AccessToken
	Users        []*User `gorm:"many2many:user_projects;"`
	Apps         []*App
}
