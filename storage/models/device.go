package models

import (
	"github.com/fsuhrau/automationhub/device"
	"gorm.io/gorm"
)

type DeviceType int

const (
	DeviceTypePhone DeviceType = iota
	DeviceTypeTablet
	DeviceTypeDesktop
	DeviceTypeUnityEditor
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
	OS               string
	OSVersion        string
	GPU              string
	ABI              string
	OpenGLESVersion  float32
	Status           device.State
	Manager          string
	StatusLog        []DeviceLog
}
