package node

import (
	"fmt"
	"sync"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/sirupsen/logrus"
)

// RPCNode is the RPC receiver on a node; the master's RPCClient calls it. The
// exported methods and their request types are the wire format.
type RPCNode struct {
	config            config.Service
	dm                manager.Devices
	connectionHandler ConnectionHandler
	abm               *AppBundleManager
	mutex             sync.Mutex
	uploadRequests    map[int32]*UploadProgress
	screenshotMutex   sync.Mutex
	screenshots       map[string][]byte
}

func NewRPCNode(config config.Service, dm manager.Devices, ch ConnectionHandler) *RPCNode {
	abm, err := NewAppBundleManager(apps.AppBundleStoragePath)
	if err != nil {
		panic(err)
	}
	return &RPCNode{
		config:            config,
		dm:                dm,
		connectionHandler: ch,
		abm:               abm,
		uploadRequests:    make(map[int32]*UploadProgress),
		screenshots:       make(map[string][]byte),
	}
}

// withDevice runs fn with the local device id, or fails when there is none.
func (s *RPCNode) withDevice(id string, fn func(device.Device) error) error {
	dev, _ := s.dm.GetDevice(id)
	if dev == nil {
		return fmt.Errorf("device with id %v not found", id)
	}
	return fn(dev)
}

// withDeviceOK is withDevice for calls that answer true once fn succeeded.
func (s *RPCNode) withDeviceOK(id string, resp *BoolResponse, fn func(device.Device) error) error {
	return s.withDevice(id, func(dev device.Device) error {
		if err := fn(dev); err != nil {
			return err
		}
		resp.Value = true
		return nil
	})
}

func (s *RPCNode) authToken() string {
	if s.config.Auth.Token == nil {
		return ""
	}
	return s.config.Auth.Token.AuthToken
}

func (s *RPCNode) DeviceStatus(req *DeviceStatusRequest, resp *DeviceResponse) error {
	logrus.Info("RPC: DeviceStatus")
	dev, mng := s.dm.GetDevice(req.DeviceId)
	if dev == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceId)
	}
	*resp = *toDeviceResponse(mng, dev)
	return nil
}

func (s *RPCNode) Ping(args *Void, reply *Void) error {
	if s.connectionHandler != nil {
		s.connectionHandler.PingReceived()
	}
	return nil
}

func (s *RPCNode) StartDevice(req *DeviceRequest, resp *ErrorResponse) error {
	logrus.Info("RPC: StartDevice")
	return s.withDevice(req.DeviceID, s.dm.Start)
}

func (s *RPCNode) StopDevice(req *DeviceRequest, resp *ErrorResponse) error {
	logrus.Info("RPC: StopDevice")
	return s.withDevice(req.DeviceID, s.dm.Stop)
}

func (s *RPCNode) GetDevices(req *Void, resp *DevicesResponse) error {
	logrus.Info("RPC: GetDevices")
	devices, err := s.dm.Devices()
	if err != nil {
		return err
	}
	resp.Devices = ToDevicesResponse(devices)
	return nil
}
