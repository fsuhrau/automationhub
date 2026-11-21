package node

import (
	"errors"

	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
)

func (s *RPCNode) IsAppInstalled(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: IsAppInstalled")
	return s.withDevice(req.DeviceID, func(dev device.Device) error {
		data, err := s.abm.GetAppParameter(req.App.Hash)
		if err != nil {
			resp.Value = false
			return nil
		}
		installed, err := dev.IsAppInstalled(getAppParameter(req, data))
		if err != nil {
			return err
		}
		resp.Value = installed
		return nil
	})
}

func (s *RPCNode) IsAppUploaded(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: IsAppUploaded")
	_, err := s.abm.GetAppParameter(req.App.Hash)
	resp.Value = !errors.Is(err, ErrAppNotFound)
	return nil
}

func (s *RPCNode) InstallApp(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: InstallApp")
	return s.withDeviceOK(req.DeviceID, resp, func(dev device.Device) error {
		data, err := s.abm.GetAppParameter(req.App.Hash)
		if err != nil {
			return err
		}
		return dev.InstallApp(getAppParameter(req, data))
	})
}

func (s *RPCNode) UninstallApp(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: UninstallApp")
	return s.withDeviceOK(req.DeviceID, resp, func(dev device.Device) error {
		return dev.UninstallApp(getAppParameter(req, nil))
	})
}

func (s *RPCNode) StartApp(req *StartAppRequest, resp *BoolResponse) error {
	logrus.Info("RPC: StartApp")
	return s.withDeviceOK(req.App.DeviceID, resp, func(dev device.Device) error {
		return dev.StartApp(getDeviceConfig(req), getAppParameter(req.App, nil), req.SessionID, s.config.NodeUrl)
	})
}

func (s *RPCNode) StopApp(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: StopApp")
	return s.withDeviceOK(req.DeviceID, resp, func(dev device.Device) error {
		if err := dev.StopApp(getAppParameter(req, nil)); err != nil {
			return err
		}
		// the master asks IsConnected next; the stopped app must not count
		device.DropAppConnection(dev)
		return nil
	})
}

func (s *RPCNode) ClearAppData(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: ClearAppData")
	return s.withDeviceOK(req.DeviceID, resp, func(dev device.Device) error {
		clearer, ok := dev.(device.AppDataClearer)
		if !ok {
			return device.ErrUnsupported
		}
		if err := clearer.ClearAppData(getAppParameter(req, nil)); err != nil {
			return err
		}
		// clearing the data kills the app as well
		device.DropAppConnection(dev)
		return nil
	})
}
