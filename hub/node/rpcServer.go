package node

import (
	"errors"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

type UploadProgress struct {
	Request  *UploadAppRequest
	received atomic.Int64
	done     atomic.Bool
	errMu    sync.Mutex
	err      error
}

func (p *UploadProgress) setError(err error) {
	p.errMu.Lock()
	p.err = err
	p.errMu.Unlock()
	logrus.Errorf("app upload %d failed: %v", p.Request.AppID, err)
}

func (p *UploadProgress) Error() error {
	p.errMu.Lock()
	defer p.errMu.Unlock()
	return p.err
}

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

func (s *RPCNode) DeviceStatus(req *DeviceStatusRequest, resp *DeviceResponse) error {
	logrus.Info("RPC: DeviceStatus")
	device, mng := s.dm.GetDevice(req.DeviceId)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceId)
	}
	*resp = *toDeviceResponse(mng, device)
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
	req, err := http.NewRequest("GET", progress.Request.URL, nil)
	if err != nil {
		progress.setError(fmt.Errorf("create request: %w", err))
		return
	}

	if s.config.Auth.Token != nil {
		req.Header.Set("X-Auth-Token", s.config.Auth.Token.AuthToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		progress.setError(fmt.Errorf("download: %w", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		progress.setError(fmt.Errorf("download: %s", resp.Status))
		return
	}

	data := make([]byte, 0, progress.Request.Size)
	buffer := make([]byte, 4*1024*1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			data = append(data, buffer[:n]...)
			progress.received.Add(int64(n))
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			progress.setError(fmt.Errorf("read file data: %w", err))
			return
		}
	}

	if progress.Request.Size > 0 && int64(len(data)) != progress.Request.Size {
		progress.setError(fmt.Errorf("size mismatch: got %d bytes, expected %d", len(data), progress.Request.Size))
		return
	}

	logrus.Infof("Download complete, storing file")
	if err := s.abm.StoreData(data, &AppBundleMetaData{
		Filename: progress.Request.Name,
		FileHash: progress.Request.Hash,
		FileSize: progress.Request.Size,
	}); err != nil {
		progress.setError(fmt.Errorf("store app bundle: %w", err))
		return
	}
	progress.done.Store(true)
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
		return fmt.Errorf("no upload in progress for app %d", req.AppID)
	}

	if err := progress.Error(); err != nil {
		delete(s.uploadRequests, req.AppID)
		return err
	}

	resp.AppID = req.AppID
	resp.DataReceived = progress.received.Load()
	if progress.done.Load() {
		// report completion only once the bundle is stored
		resp.DataReceived = progress.Request.Size
		delete(s.uploadRequests, req.AppID)
	} else if resp.DataReceived >= progress.Request.Size {
		// downloaded but not stored yet
		resp.DataReceived = progress.Request.Size - 1
	}

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

func getDeviceConfig(req *StartAppRequest) *device.DeviceConfig {
	var deviceConfig device.DeviceConfig
	if req.ConnectionParams != nil {
		deviceConfig.Connection = int(req.ConnectionParams.Type)
		deviceConfig.IP = req.ConnectionParams.IP
		deviceConfig.Port = int(req.ConnectionParams.Port)
	}
	deviceConfig.DeviceParameter = make(map[string]string)
	for _, param := range req.CustomParameter {
		deviceConfig.DeviceParameter[param.Key] = param.Value
	}
	return &deviceConfig
}

func (s *RPCNode) StartApp(req *StartAppRequest, resp *BoolResponse) error {
	logrus.Info("RPC: StartApp")
	device, _ := s.dm.GetDevice(req.App.DeviceID)
	if device == nil {
		return fmt.Errorf("device with id %v not found", req.App.DeviceID)
	}

	deviceConfig := getDeviceConfig(req)

	err := device.StartApp(deviceConfig, getAppParameter(req.App, nil), req.SessionID, s.config.NodeUrl)

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

func (s *RPCNode) ClearAppData(req *AppParameterRequest, resp *BoolResponse) error {
	logrus.Info("RPC: ClearAppData")
	dev, _ := s.dm.GetDevice(req.DeviceID)
	if dev == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}
	clearer, ok := dev.(device.AppDataClearer)
	if !ok {
		return device.ErrUnsupported
	}
	if err := clearer.ClearAppData(getAppParameter(req, nil)); err != nil {
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
	dev, _ := s.dm.GetDevice(req.DeviceID)
	if dev == nil {
		return fmt.Errorf("device with id %v not found", req.DeviceID)
	}

	b, w, h, err := dev.GetScreenshot()
	if err != nil {
		return err
	}

	resp.Hash = req.DeviceID
	resp.Size = int32(len(b))
	resp.Width = int32(w)
	resp.Height = int32(h)

	s.screenshotMutex.Lock()
	s.screenshots[req.DeviceID] = b
	s.screenshotMutex.Unlock()
	return nil
}

// GetScreenshotData returns a slice of the last screenshot taken for the
// device. The buffer is released once its last chunk was requested.
func (s *RPCNode) GetScreenshotData(req *ScreenShotDataRequest, resp *ScreenShotDataResponse) error {
	s.screenshotMutex.Lock()
	defer s.screenshotMutex.Unlock()
	data, ok := s.screenshots[req.Hash]
	if !ok {
		return fmt.Errorf("no screenshot data for %s", req.Hash)
	}
	start, end := int(req.Start), int(req.End)
	if start < 0 || end < start || start > len(data) {
		return fmt.Errorf("invalid screenshot range %d-%d (size %d)", start, end, len(data))
	}
	if end > len(data) {
		end = len(data)
	}
	resp.Data = data[start:end]
	if end == len(data) {
		delete(s.screenshots, req.Hash)
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
