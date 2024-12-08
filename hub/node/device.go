package node

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/hub/manager"
	"net"
	"time"
)

type RPCDevice struct {
	generic.Device
	Dev        *DeviceResponse
	client     manager.RPCClient
	connection *device.Connection
}

func (d *RPCDevice) DeviceID() string {
	return d.Dev.DeviceID
}

func (d *RPCDevice) DeviceIP() net.IP {
	return net.ParseIP(d.Dev.DeviceIP)
}

func (d *RPCDevice) DeviceState() device.State {
	return toDeviceState(d.Dev.State)
}

func (d *RPCDevice) DeviceOSName() string {
	return d.Dev.DeviceOS
}

func (d *RPCDevice) DeviceOSVersion() string {
	return d.Dev.DeviceOSVersion
}

func (d *RPCDevice) DeviceName() string {
	return d.Dev.Name
}

func (d *RPCDevice) DeviceModel() string {
	return d.Dev.DeviceModel
}

func (d *RPCDevice) IsAppInstalled(*app.Parameter) (bool, error) {
	return true, nil
}

func (d *RPCDevice) InstallApp(*app.Parameter) error {
	return nil
}

func (d *RPCDevice) UninstallApp(*app.Parameter) error {
	return nil
}

func (d *RPCDevice) UpdateDeviceInfos() error {
	return nil

}

func (d *RPCDevice) ConnectionTimeout() time.Duration {
	return 30 * time.Second
}

func (d *RPCDevice) IsAppConnected() bool {
	return true
}

func (d *RPCDevice) StartApp(*app.Parameter, string, string) error {
	return nil
}

func (d *RPCDevice) StopApp(*app.Parameter) error {
	return nil
}

func (d *RPCDevice) StartRecording(string) error {
	return nil
}

func (d *RPCDevice) StopRecording() error {
	return nil
}

func (d *RPCDevice) GetScreenshot() ([]byte, int, int, error) {
	return nil, 0, 0, nil
}

func (d *RPCDevice) HasFeature(string) bool {
	return true
}

func (d *RPCDevice) Execute(string) {

}
func (d *RPCDevice) RunNativeScript(data []byte) {

}
