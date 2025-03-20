package device

import (
	"github.com/gorilla/websocket"
	"net"
)

type RegisterData struct {
	ManagerType     string
	DeviceID        string
	Name            string
	DeviceOS        string
	DeviceOSVersion string
	TargetVersion   string
	DeviceOSInfos   string
	ProjectDir      string
	DeviceModel     string
	RAM             float32
	GPU             string
	SOC             string
	DeviceIP        net.IP
	Conn            *websocket.Conn
}
