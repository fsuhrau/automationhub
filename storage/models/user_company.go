package models

import "gorm.io/gorm"

type UserCompany struct{
	gorm.Model
	UserID uint
	CompanyID uint
}