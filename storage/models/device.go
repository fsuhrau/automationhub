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

type ConnectionType int

const (
	ConnectionTypeUSB ConnectionType = iota
	ConnectionTypeRemote
)

type Devices []*Device

type Device struct {
	gorm.Model
	CompanyID           uint
	DeviceIdentifier    string
	DeviceType          DeviceType
	Name                string
	Manager             string
	HardwareModel       string
	RAM                 float32
	SOC                 string
	DisplaySize         string
	DPI                 float32
	OS                  string
	OSVersion           string
	GPU                 string
	ABI                 string
	OpenGLESVersion     float32
	Parameter           []DeviceParameter
	StatusLog           []DeviceLog
	ConnectionParameter ConnectionParameter
	IsAcknowledged      bool
	Status              device.State `gorm:"-"`
	Dev                 interface{}  `gorm:"-"`
	Connection          interface{}  `gorm:"-"`
}

type ConnectionParameter struct {
	gorm.Model
	DeviceID       uint
	ConnectionType ConnectionType
	IP             string
	Port           int
}

type DeviceParameter struct {
	gorm.Model
	DeviceID uint
	Key      string
	Value    string
}

func (d *Device) GetAttribute(key string) string {
	for i := range d.Parameter {
		if d.Parameter[i].Key == key {
			return d.Parameter[i].Value
		}
	}
	return ""
}
