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
	DeviceOSInfos   string
	DeviceIP        net.IP
	Conn            *websocket.Conn
}
