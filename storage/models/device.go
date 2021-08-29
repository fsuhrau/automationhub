package models

import (
	"gorm.io/gorm"
)

type DeviceType int

type DeviceStatus int

type Device struct {
	gorm.Model
	CompanyID uint
	DeviceIdentifier string
	DeviceType DeviceType
	Name string
	RAM float32
	SOC string
	DisplaySize string
	DPI float32
	OSVersion string
	GPU string
	ABI string
	OpenGLESVersion float32
	Status DeviceStatus
	Manager string
}