package devices

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/inspector/handler/manager"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
)

type dev struct {
	Identifier             string
	Name                   string
	OperationSystem        string
	SupportedArchitectures string
	Status                 string
}

type devs []*dev
func (s devs) Len() int      { return len(s) }
func (s devs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByName struct{ devs }
func (s ByName) Less(i, j int) bool { return s.devs[i].Name < s.devs[j].Name }

func Index(manager manager.DeviceManager) func(*gin.Context) {
	return func(c *gin.Context) {

		var deviceList []*dev
		devices, _ := manager.Devices()
		for _, d := range devices {
			deviceList = append(deviceList, &dev{
				Identifier:   d.DeviceID(),
				Name: d.DeviceName(),
				OperationSystem:   d.DeviceOSName(),
				SupportedArchitectures: "",
				Status: device.StateToString(d.DeviceState()),
			})
		}

		sort.Sort(ByName{deviceList})

		c.HTML(http.StatusOK, "devices/index", gin.H{
			"devices": deviceList,
		})
	}
}
