package dashboard

import (
	"net/http"

	"github.com/fsuhrau/automationhub/inspector/handler/manager"
	"github.com/gin-gonic/gin"
)

func Index(dm manager.DeviceManager, sm manager.SessionManager) func(*gin.Context) {
	return func(c *gin.Context) {
		deviceList := manager.DeviceList(dm)
		var byOs map[string][]*manager.Dev
		byOs = make(map[string][]*manager.Dev)
		for _, device := range deviceList {
			var list []*manager.Dev
			var ok bool
			if list, ok = byOs[device.OperationSystem]; ok {
				list = append(list, device)
			} else {
				list = append(list, device)
			}
			byOs[device.OperationSystem] = list
		}

		var numLockedAndroid int
		var numLockedIOS int
		var numLockedSimulator int
		var numLockedMac int
		sessions := sm.GetSessions()
		for _, s := range sessions {
			device := s.GetDevice()
			if device != nil {
				switch device.DeviceOSName() {
				case "android":
					numLockedAndroid++
				case "iphoneos":
					numLockedIOS++
				case "iphonesimulator":
					numLockedSimulator++
				case "MacOSX":
					numLockedMac++
				}
			}
		}

		var numAndroid int
		var numIOS int
		var numSimulator int
		var numMac int
		if v, ok := byOs["android"]; ok {
			numAndroid = len(v)
		}
		if v, ok := byOs["iphoneos"]; ok {
			numIOS = len(v)
		}
		if v, ok := byOs["iphonesimulator"]; ok {
			numSimulator = len(v)
		}
		if v, ok := byOs["MacOSX"]; ok {
			numMac = len(v)
		}

		c.HTML(http.StatusOK, "dashboard/index", gin.H{
			"numIOS":           numIOS,
			"numAndroid":       numAndroid,
			"numSimulator":     numSimulator,
			"numMac":           numMac,
			"numIOSUsed":       numLockedIOS,
			"numAndroidUsed":   numLockedAndroid,
			"numSimulatorUsed": numLockedSimulator,
			"numMacUsed":       numLockedMac,
			"sessions":         len(sessions),
			"failedTests":      0,
			"succeededTests":   0,
			"skipedTests":      0,
		})
	}
}
