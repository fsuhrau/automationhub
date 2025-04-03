package node

import (
	"encoding/json"
	device2 "github.com/fsuhrau/automationhub/device"
)

func fromDeviceState(state device2.State) DeviceState {
	switch state {
	case device2.StateNone:
		return DeviceState_None
	case device2.StateUnknown:
		return DeviceState_Unknown
	case device2.StateShutdown:
		return DeviceState_Shutdown
	case device2.StateRemoteDisconnected:
		return DeviceState_RemoteDisconnected
	case device2.StateBooted:
		return DeviceState_Booted
	case device2.StateLocked:
		return DeviceState_Locked
	}
	return DeviceState_Unknown
}

func toDeviceState(state DeviceState) device2.State {
	switch state {
	case DeviceState_None:
		return device2.StateNone
	case DeviceState_Unknown:
		return device2.StateUnknown
	case DeviceState_Shutdown:
		return device2.StateShutdown
	case DeviceState_RemoteDisconnected:
		return device2.StateRemoteDisconnected
	case DeviceState_Booted:
		return device2.StateBooted
	case DeviceState_Locked:
		return device2.StateLocked
	}
	return device2.StateUnknown
}

func toDeviceResponse(manager string, dev device2.Device) *DeviceResponse {
	data, _ := json.Marshal(dev.DeviceParameter())
	return &DeviceResponse{
		Manager:         manager,
		DeviceID:        dev.DeviceID(),
		Name:            dev.DeviceName(),
		DeviceOS:        dev.DeviceOSName(),
		DeviceOSVersion: dev.DeviceOSVersion(),
		ProjectDir:      "",
		State:           fromDeviceState(dev.DeviceState()),
		DeviceType:      int32(dev.DeviceType()),
		DeviceParameter: string(data),
		PlatformType:    int32(dev.PlatformType()),
	}
}
