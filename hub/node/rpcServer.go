package node

import (
	"errors"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sync"
	"time"
)

type UploadProgress struct {
	Request           *UploadAppRequest
	NextProgressPrint *time.Time
	Data              []byte
}

type RPCNode struct {
	config            config.Service
	dm                manager.Devices
	connectionHandler ConnectionHandler
	abm               *AppBundleManager
	mutex             sync.Mutex
	uploadRequests    map[int32]*UploadProgress
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

func (s *RPCNode) DeviceStatus(req *DeviceStatusRequest, resp *DeviceResponse) error {
	logrus.Info("RPC: DeviceStatus")
	device, mng := s.dm.GetDevice(req.DeviceId)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceId)
	}
	resp = toDeviceResponse(mng, device)
	return nil
}

func (s *RPCNode) Ping(args *Void, reply *Void) error {
	// logrus.Info("RPC: Ping")
	if s.connectionHandler != nil {
		s.connectionHandler.PingReceived()
	}
	return nil
}

func (s *RPCNode) StartDevice(req *DeviceRequest, resp *ErrorResponse) error {
	logrus.Info("RPC: StartDevice")
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
	logrus.Info("RPC: StopDevice")
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
	logrus.Info("RPC: GetDevices")
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

func getAppParameter(req *AppParameterRequest, data *AppBundleMetaData) *app.Parameter {
	appPath := ""
	if data != nil {
		appPath = data.FilePath
	}

	var appParams *app.AppParams
	if req.App != nil {
		var androidParams *app.AndroidParams
		if req.App.Android != nil {
			androidParams = &app.AndroidParams{
				LaunchActivity: req.App.Android.LaunchActivity,
			}
		}

		var executableParams *app.ExecutableParams
		if req.App.Executable != nil {
			executableParams = &app.ExecutableParams{
				Executable: req.App.Executable.Executable,
			}
		}
		appParams = &app.AppParams{
			AppPath:    appPath,
			Additional: req.App.Additional,
			Hash:       req.App.Hash,
			Size:       int(req.App.Size),
			Android:    androidParams,
			Executable: executableParams,
		}
	}
	var webParams *app.WebParams
	if req.Web != nil {
		webParams = &app.WebParams{
			StartURL: req.Web.StartURL,
		}
	}

	return &app.Parameter{
		Platform:   req.Platform,
		Name:       req.Name,
		Identifier: req.Identifier,
		Version:    req.Version,
		App:        appParams,
		Web:        webParams,
	}
}

func (s *RPCNode) IsAppInstalled(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: IsAppInstalled")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	data, err := s.abm.GetAppParameter(req.App.Hash)
	if err != nil {
		resp.Value = false
		return nil
	}

	installed, err := device.IsAppInstalled(getAppParameter(req, data))

	if err != nil {
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = 1
		return err
	}
	resp.Value = installed
	return nil
}

func (s *RPCNode) IsAppUploaded(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: IsAppUploaded")

	if _, err := s.abm.GetAppParameter(req.App.Hash); errors.Is(err, ErrAppNotFound) {
		resp.Value = false
		return nil
	}

	resp.Value = true
	return nil
}

func (s *RPCNode) UploadApp(req *UploadAppRequest, resp *BoolResponse) error {
	logrus.Info("RPC: UploadApp")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	progress := &UploadProgress{
		Request: req,
	}

	s.uploadRequests[req.AppID] = progress

	go s.downloadFile(progress)

	resp.Value = true
	return nil
}

func (s *RPCNode) downloadFile(progress *UploadProgress) {
	resp, err := http.Get(progress.Request.URL)
	if err != nil {
		logrus.Error("Failed to download file: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Error("Failed to download file: ", resp.Status)
		return
	}

	progress.Data = make([]byte, 0, progress.Request.Size)
	buffer := make([]byte, 10*1024*1024) // 10MB buffer size
	totalRead := int64(0)

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			progress.Data = append(progress.Data, buffer[:n]...)
			totalRead += int64(n)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			logrus.Error("Failed to read file data: ", err)
			return
		}
	}

	logrus.Infof("Download Complete stroging file")
	if err := s.abm.StoreData(progress.Data, &AppBundleMetaData{
		Filename: progress.Request.Name,
		FileHash: progress.Request.Hash,
		FileSize: progress.Request.Size,
	}); err != nil {
		logrus.Error(err)
	}
}

func humanReadableSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (s *RPCNode) UploadAppProgress(req *UploadAppProgressRequest, resp *UploadAppProgressResponse) error {
	// logrus.Info("RPC: UploadApp")

	s.mutex.Lock()
	defer s.mutex.Unlock()

	progress, exists := s.uploadRequests[req.AppID]
	if !exists {
		err := fmt.Errorf("No Meta Progress found for Upload Data")
		return err
	}

	resp.AppID = req.AppID
	resp.DataReceived = int64(len(progress.Data))

	return nil
}

func (s *RPCNode) InstallApp(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: InstallApp")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	data, err := s.abm.GetAppParameter(req.App.Hash)
	if err != nil {
		return err
	}

	err = device.InstallApp(getAppParameter(req, data))

	if err != nil {
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = 1
		return err
	}
	resp.Value = true
	return nil
}

func (s *RPCNode) UninstallApp(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: UninstallApp")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	err := device.UninstallApp(getAppParameter(req, nil))

	if err != nil {
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = 1
		return err
	}
	resp.Value = true
	return nil
}

func (s *RPCNode) StartApp(req *StartAppRequest, resp *BoolResponse) error {
	logrus.Info("RPC: StartApp")
	device, _ := s.dm.GetDevice(req.App.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.App.DeviceID)
	}

	err := device.StartApp(getAppParameter(req.App, nil), req.SessionID, s.config.NodeUrl)

	if err != nil {
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = 1
		return err
	}
	resp.Value = true
	return nil
}

func (s *RPCNode) StopApp(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: StopApp")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	err := device.StopApp(getAppParameter(req, nil))

	if err != nil {
		resp.ErrorMessage = err.Error()
		resp.ErrorCode = 1
		return err
	}
	resp.Value = true
	return nil
}

func (s *RPCNode) IsConnected(req *DeviceRequest, resp *BoolResponse) error {
	logrus.Info("RPC: IsConnected")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	connected := device.IsAppConnected()

	resp.Value = connected
	return nil
}

func (s *RPCNode) TakeScreenshot(req *DeviceRequest, resp *ScreenShotResponse) error {
	logrus.Info("RPC: TakeScreenshot")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	b, w, h, err := device.GetScreenshot()
	resp.Hash = req.DeviceID
	resp.Size = int32(len(b))
	resp.Width = int32(w)
	resp.Height = int32(h)

	s.screenshots[req.DeviceID] = b
	return err
}

func (s *RPCNode) GetScreenshotData(req *ScreenShotDataRequest, resp *ScreenShotDataResponse) error {
	if data, ok := s.screenshots[req.Hash]; ok {
		resp.Data = data[req.Start:req.End]
	}
	return nil
}

func (s *RPCNode) HasFeature(req *FeatureRequest, resp *BoolResponse) error {
	logrus.Info("RPC: HasFeature")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	supported := device.HasFeature(req.Feature)

	resp.Value = supported
	return nil
}

func (s *RPCNode) Execute(req *ExecuteRequest, resp *Void) error {
	logrus.Info("RPC: Execute")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	device.Execute(req.Data)

	return nil
}

func (s *RPCNode) ConnectionTimeout(req *DeviceRequest, resp *TimeoutResponse) error {
	logrus.Info("RPC: ConnectionTimeout")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	timeout := device.ConnectionTimeout()
	resp.Timeout = int64(timeout)
	return nil
}

func (s *RPCNode) RunNativeScript(req *ExecuteRequest, resp *Void) error {
	logrus.Info("RPC: RunNativeScript")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	device.RunNativeScript([]byte(req.Data))
	return nil
}

func (s *RPCNode) SendAction(req *ExecuteRequest, resp *Void) error {
	logrus.Info("RPC: SendAction")
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	return device.Send([]byte(req.Data))
}
