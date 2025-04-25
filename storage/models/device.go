package models

import (
	"github.com/fsuhrau/automationhub/device"
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
	Model
	CompanyID           uint                 `json:"companyId"`
	DeviceIdentifier    string               `json:"deviceIdentifier"`
	Alias               string               `json:"alias"`
	NodeID              uint                 `json:"nodeId"`
	Node                *Node                `json:"node"`
	DeviceType          DeviceType           `json:"deviceType"`
	Name                string               `json:"name"`
	Manager             string               `json:"manager"`
	OS                  string               `json:"os"`
	OSVersion           string               `json:"osVersion"`
	OSInfos             string               `json:"osInfos"`
	TargetVersion       string               `json:"targetVersion"`
	DeviceParameter     []DeviceParameter    `json:"deviceParameter"`
	StatusLog           []DeviceLog          `json:"statusLog"`
	ConnectionType      ConnectionType       `json:"connectionType"`
	ConnectionParameter *ConnectionParameter `json:"connectionParameter"`
	IsAcknowledged      bool                 `json:"isAcknowledged"`
	Status              device.State         `json:"status" gorm:"-"`
	IsLocked            bool                 `json:"isLocked" gorm:"-"`
	Dev                 interface{}          `json:"dev" gorm:"-"`
	Connection          interface{}          `json:"connection" gorm:"-"`
	PlatformType        PlatformType         `json:"platformType"`
	CustomParameter     []CustomParameter    `json:"customParameter"`
}

type ConnectionParameter struct {
	Model
	DeviceID       uint           `json:"deviceId"`
	Device         *Device        `json:"device"`
	ConnectionType ConnectionType `json:"connectionType"`
	IP             string         `json:"ip"`
	Port           int            `json:"port"`
}

type DeviceParameter struct {
	Model
	DeviceID uint    `json:"deviceId"`
	Device   *Device `json:"device"`
	Key      string  `json:"key"`
	Value    string  `json:"value"`
}

type CustomParameter struct {
	Model
	DeviceID uint    `json:"deviceId"`
	Device   *Device `json:"device"`
	Key      string  `json:"key"`
	Value    string  `json:"value"`
}

func (d *Device) GetAttribute(key string) string {
	for i := range d.CustomParameter {
		if d.CustomParameter[i].Key == key {
			return d.CustomParameter[i].Value
		}
	}
	return ""
}
