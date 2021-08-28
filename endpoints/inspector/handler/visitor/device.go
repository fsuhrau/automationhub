package visitor

import (
	"sort"

	"github.com/fsuhrau/automationhub/device"
)

type DeviceManager interface {
	Devices() ([]device.Device, error)
}

type Dev struct {
	Identifier             string
	Name                   string
	OperationSystem        string
	OperationSystemVersion string
	SupportedArchitectures string
	Status                 string
}

type Devs []*Dev

func (s Devs) Len() int      { return len(s) }
func (s Devs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByName struct{ Devs }

func (s ByName) Less(i, j int) bool { return s.Devs[i].Name < s.Devs[j].Name }

func DeviceList(m DeviceManager) (deviceList []*Dev) {
	devices, _ := m.Devices()
	for _, d := range devices {
		deviceList = append(deviceList, &Dev{
			Identifier:             d.DeviceID(),
			Name:                   d.DeviceName(),
			OperationSystem:        d.DeviceOSName(),
			OperationSystemVersion: d.DeviceOSVersion(),
			SupportedArchitectures: "",
			Status:                 device.StateToString(d.DeviceState()),
		})
	}
	sort.Sort(ByName{deviceList})
	return
}
