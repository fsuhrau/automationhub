package node

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/manager"
)

type RPCNode struct {
	config            config.Service
	dm                manager.Devices
	connectionHandler ConnectionHandler
}

func NewRPCNode(config config.Service, dm manager.Devices, ch ConnectionHandler) *RPCNode {
	return &RPCNode{
		config:            config,
		dm:                dm,
		connectionHandler: ch,
	}
}

func (s *RPCNode) DeviceStatus(req *DeviceStatusRequest, resp *DeviceResponse) error {
	fmt.Println("RPC: DeviceStatus")
	device, mng := s.dm.GetDevice(req.DeviceId)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceId)
	}
	resp = toDeviceResponse(mng, device)
	return nil
}

func (s *RPCNode) Ping(args *Void, reply *Void) error {
	// fmt.Println("RPC: Ping")
	if s.connectionHandler != nil {
		s.connectionHandler.PingReceived()
	}
	return nil
}

func (s *RPCNode) StartDevice(req *DeviceRequest, resp *ErrorResponse) error {
	fmt.Println("RPC: StartDevice")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}
	if err := s.dm.Start(device); err != nil {
		return err
	}
	return nil
}

func (s *RPCNode) StopDevice(req *DeviceRequest, resp *ErrorResponse) error {
	fmt.Println("RPC: StopDevice")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}
	if err := s.dm.Stop(device); err != nil {
		return err
	}
	return nil
}

func (s *RPCNode) GetDevices(req *Void, resp *DevicesResponse) error {
	fmt.Println("RPC: GetDevices")
	devices, err := s.dm.Devices()
	if err != nil {
		return err
	}

	for m, devs := range devices {
		for _, d := range devs {
			resp.Devices = append(resp.Devices, toDeviceResponse(m, d))
		}
	}

	return nil
}

func getAppParameter(req *AppParameterRequest) *app.Parameter {
	return &app.Parameter{
		AppBinaryID:    uint(req.AppID),
		Platform:       req.Platform,
		Name:           req.Name,
		Identifier:     req.Identifier,
		Version:        req.Version,
		LaunchActivity: req.LaunchActivity,
		Additional:     req.Additional,
		Hash:           req.Hash,
		Size:           int(req.Size),
	}
}

func (s *RPCNode) IsAppInstalled(req *AppParameterRequest, resp *BoolResponse) error {
	fmt.Println("RPC: IsAppInstalled")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	installed, err := device.IsAppInstalled(getAppParameter(req))

	if err != nil {
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = 1
		return err
	}
	resp.Value = installed
	return nil
}

func (s *RPCNode) UploadApp(req *UploadAppRequest, resp *BoolResponse) error {
	fmt.Println("RPC: UploadApp")
	// needs to be implemented still
	resp.Value = true
	return nil
}

func (s *RPCNode) InstallApp(req *AppParameterRequest, resp *BoolResponse) error {
	fmt.Println("RPC: InstallApp")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	err := device.InstallApp(getAppParameter(req))

	if err != nil {
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = 1
		return err
	}
	resp.Value = true
	return nil
}

func (s *RPCNode) UninstallApp(req *AppParameterRequest, resp *BoolResponse) error {
	fmt.Println("RPC: UninstallApp")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	err := device.UninstallApp(getAppParameter(req))

	if err != nil {
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = 1
		return err
	}
	resp.Value = true
	return nil
}

func (s *RPCNode) StartApp(req *StartAppRequest, resp *BoolResponse) error {
	fmt.Println("RPC: StartApp")
	device, _ := s.dm.GetDevice(req.App.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.App.DeviceID)
	}

	err := device.StartApp(getAppParameter(req.App), req.SessionID, s.config.NodeUrl)

	if err != nil {
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = 1
		return err
	}
	resp.Value = true
	return nil
}

func (s *RPCNode) StopApp(req *AppParameterRequest, resp *BoolResponse) error {
	fmt.Println("RPC: StopApp")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	err := device.StopApp(getAppParameter(req))

	if err != nil {
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = 1
		return err
	}
	resp.Value = true
	return nil
}

func (s *RPCNode) IsConnected(req *DeviceRequest, resp *BoolResponse) error {
	fmt.Println("RPC: IsConnected")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	connected := device.IsAppConnected()

	resp.Value = connected
	return nil
}

func (s *RPCNode) GetScreenshot(req *DeviceRequest, resp *ScreenShotResponse) error {
	fmt.Println("RPC: GetScreenshot")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	b, w, h, err := device.GetScreenshot()
	resp.Data = b
	resp.Width = int32(w)
	resp.Height = int32(h)
	return err
}

func (s *RPCNode) HasFeature(req *FeatureRequest, resp *BoolResponse) error {
	fmt.Println("RPC: HasFeature")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	supported := device.HasFeature(req.Feature)

	resp.Value = supported
	return nil
}

func (s *RPCNode) Execute(req *ExecuteRequest, resp *Void) error {
	fmt.Println("RPC: Execute")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	device.Execute(req.Data)

	return nil
}

func (s *RPCNode) ConnectionTimeout(req *DeviceRequest, resp *TimeoutResponse) error {
	fmt.Println("RPC: ConnectionTimeout")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	timeout := device.ConnectionTimeout()
	resp.Timeout = int64(timeout)
	return nil
}

func (s *RPCNode) RunNativeScript(req *ExecuteRequest, resp *Void) error {
	fmt.Println("RPC: RunNativeScript")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	device.RunNativeScript([]byte(req.Data))
	return nil
}

func (s *RPCNode) SendAction(req *ExecuteRequest, resp *Void) error {
	fmt.Println("RPC: SendAction")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	return device.Send([]byte(req.Data))
}
