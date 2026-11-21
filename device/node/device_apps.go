package node

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

func (d *NodeDevice) IsAppInstalled(params *app.Parameter) (bool, error) {
	return d.nodeManager.IsAppInstalled(d.nodeId, d.ID, params)
}

func (d *NodeDevice) InstallApp(params *app.Parameter) error {
	return d.nodeManager.InstallApp(d.nodeId, d.ID, params)
}

func (d *NodeDevice) UninstallApp(params *app.Parameter) error {
	return d.nodeManager.UninstallApp(d.nodeId, d.ID, params)
}

// StartApp starts the app on the node with the connection settings stored for
// the device; the config argument of the caller does not apply to node devices.
//
//nolint:revive // argument-limit: fixed by device.AppManager
func (d *NodeDevice) StartApp(_ *device.DeviceConfig, appParams *app.Parameter, sessionId string, nodeUrl string) error {
	return d.nodeManager.StartApp(d.nodeId, d.ID, d.storedDeviceConfig(), appParams, sessionId, nodeUrl)
}

// storedDeviceConfig is the connection and custom parameters configured for
// the device in the database.
func (d *NodeDevice) storedDeviceConfig() *device.DeviceConfig {
	config := &device.DeviceConfig{DeviceParameter: make(map[string]string)}
	stored := d.Config
	if stored == nil {
		return config
	}
	if stored.ConnectionParameter != nil {
		config.Connection = int(stored.ConnectionParameter.ConnectionType)
		config.IP = stored.ConnectionParameter.IP
		config.Port = stored.ConnectionParameter.Port
	}
	for _, param := range stored.CustomParameter {
		config.DeviceParameter[param.Key] = param.Value
	}
	return config
}

func (d *NodeDevice) StopApp(params *app.Parameter) error {
	return d.nodeManager.StopApp(d.nodeId, d.ID, params)
}

func (d *NodeDevice) ClearAppData(params *app.Parameter) error {
	return d.nodeManager.ClearAppData(d.nodeId, d.ID, params)
}
