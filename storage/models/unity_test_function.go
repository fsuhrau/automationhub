package models

import "gorm.io/gorm"

type UnityTestFunction struct {
	gorm.Model
	TestConfigUnityID uint
	Assembly          string
	Class             string
	Method            string
}
