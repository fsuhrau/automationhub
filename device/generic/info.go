package generic

import "github.com/fsuhrau/automationhub/device"

// Info holds the identity of a device and implements device.Identity. Platform
// devices embed it.
type Info struct {
	ID        string
	Name      string
	OSName    string
	OSVersion string
	State     device.State
	Type      int
	Platform  int
	Parameter map[string]string
	// Target is the version the app is built for, see TargetVersion.
	Target string
}

var _ device.Identity = Info{}

func (i Info) DeviceID() string                   { return i.ID }
func (i Info) DeviceName() string                 { return i.Name }
func (i Info) DeviceOSName() string               { return i.OSName }
func (i Info) DeviceOSVersion() string            { return i.OSVersion }
func (i Info) DeviceState() device.State          { return i.State }
func (i Info) DeviceType() int                    { return i.Type }
func (i Info) PlatformType() int                  { return i.Platform }
func (i Info) DeviceParameter() map[string]string { return i.Parameter }
func (i Info) TargetVersion() string              { return i.Target }
