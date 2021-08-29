package models

import (
	"gorm.io/gorm"
)

type DeviceType int
const (
	DeviceType_Phone DeviceType = iota
	DeviceType_Tablet
	DeviceType_MacOS
	DeviceType_Windows
	DeviceType_Unity
)

type DeviceStatus int
const (
	DeviceStatus_Unknown            DeviceStatus = iota
	DeviceStatus_Shutdown
	DeviceStatus_RemoteDisconnected
	DeviceStatus_Booted
	DeviceStatus_Locked
	DeviceStatus_Unlocked
)

type Device struct {
	gorm.Model
	CompanyID        uint
	DeviceIdentifier string
	DeviceType       DeviceType
	Name             string
	RAM              float32
	SOC              string
	DisplaySize      string
	DPI              float32
	OSVersion        string
	GPU              string
	ABI              string
	OpenGLESVersion  float32
	Status           DeviceStatus
	Manager          string
	StatusLog        []DeviceLog
}
