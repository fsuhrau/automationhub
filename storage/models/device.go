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
	ConnectionTypeNode
)

type PlatformType uint

const (
	PlatformTypeiOS PlatformType = iota
	PlatformTypeAndroid
	PlatformTypeMac
	PlatformTypeWindows
	PlatformTypeLinux
	PlatformTypeWeb
	PlatformTypeEditor
	PlatformTypeiOSSimulator
)

type Devices []*Device

type Device struct {
	gorm.Model
	CompanyID           uint
	DeviceIdentifier    string
	Alias               string
	NodeID              uint
	Node                *Node
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
	OSInfos             string
	GPU                 string
	ABI                 string
	TargetVersion       string
	OpenGLESVersion     float32
	Parameter           []DeviceParameter
	StatusLog           []DeviceLog
	ConnectionParameter *ConnectionParameter
	IsAcknowledged      bool
	Status              device.State `gorm:"-"`
	Dev                 interface{}  `gorm:"-"`
	Connection          interface{}  `gorm:"-"`
	PlatformType        PlatformType
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
