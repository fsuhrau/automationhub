package models

import "gorm.io/gorm"

type TestType uint

type TestConfig struct {
	gorm.Model
	TestID uint
	Type   TestType
}
