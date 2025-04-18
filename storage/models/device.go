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
	DeviceTypeBrowser
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
	OS                  string
	OSVersion           string
	OSInfos             string
	TargetVersion       string
	DeviceParameter     []DeviceParameter
	StatusLog           []DeviceLog
	ConnectionType      ConnectionType
	ConnectionParameter *ConnectionParameter
	IsAcknowledged      bool
	Status              device.State `gorm:"-"`
	IsLocked            bool         `gorm:"-"`
	Dev                 interface{}  `gorm:"-"`
	Connection          interface{}  `gorm:"-"`
	PlatformType        PlatformType
	CustomParameter     []CustomParameter
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

type CustomParameter struct {
	gorm.Model
	DeviceID uint
	Key      string
	Value    string
}

func (d *Device) GetAttribute(key string) string {
	for i := range d.CustomParameter {
		if d.CustomParameter[i].Key == key {
			return d.CustomParameter[i].Value
		}
	}
	return ""
}
