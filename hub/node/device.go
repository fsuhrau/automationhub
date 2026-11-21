package node

import (
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"net"
	"time"
)

type RPCDevice struct {
	generic.Device
	Dev *DeviceResponse
}

func (d *RPCDevice) PlatformType() int {
	return int(d.Dev.PlatformType)
}

func (d *RPCDevice) DeviceParameter() map[string]string {
	deviceParams := make(map[string]string)
	_ = json.Unmarshal([]byte(d.Dev.DeviceParameter), &deviceParams)
	return deviceParams
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

func (d *RPCDevice) TargetVersion() string {
	return d.Dev.TargetVersion
}

func (d *RPCDevice) DeviceName() string {
	return d.Dev.Name
}

func (d *RPCDevice) DeviceType() int {
	return int(d.Dev.DeviceType)
}

// RPCDevice only describes a device of a node. It can't be controlled: the
// master wraps it into a device/node.NodeDevice, which forwards to the node.
// Its app management says so instead of pretending to work.

func (d *RPCDevice) IsAppInstalled(*app.Parameter) (bool, error) {
	return false, errDescriptionOnly("is app installed")
}

func (d *RPCDevice) InstallApp(*app.Parameter) error {
	return errDescriptionOnly("install app")
}

func (d *RPCDevice) UninstallApp(*app.Parameter) error {
	return errDescriptionOnly("uninstall app")
}

func (d *RPCDevice) StartApp(*device.DeviceConfig, *app.Parameter, string, string) error {
	return errDescriptionOnly("start app")
}

func (d *RPCDevice) StopApp(*app.Parameter) error {
	return errDescriptionOnly("stop app")
}

func (d *RPCDevice) ConnectionTimeout() time.Duration {
	return 30 * time.Second
}

// IsAppConnected is false: the description has no connection to the app.
func (d *RPCDevice) IsAppConnected() bool {
	return false
}

func errDescriptionOnly(op string) error {
	return fmt.Errorf("%s: %w", op, device.ErrUnsupported)
}
