package manager

import (
	"github.com/fsuhrau/automationhub/device"
	"net"
	"sync"
)

type ResponseData struct {
	Data []byte
	Err  error
}

type DeviceLock struct {
	Device          device.Device
	Connection      net.Conn
	AppName         string
	WaitingGroup    *sync.WaitGroup
	ResponseChannel chan ResponseData
	//ConnectionStateChannel chan bool
}