package models

import (
	"github.com/fsuhrau/automationhub/device"
	"time"
)

type DeviceLog struct {
	Model
	DeviceID uint         `json:"deviceId"`
	Device   *Device      `json:"device"`
	Time     time.Time    `json:"time"`
	Status   device.State `json:"status"`
	Payload  string       `json:"payload"`
}
