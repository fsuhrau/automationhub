package node

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/config/protocol"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/node/jsonrpc"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/rpc"
	"net/url"
	"path/filepath"
	"time"
)

type RPCClient struct {
	client    *rpc.Client
	masterURL string
}

func NewRPCClient(conn *websocket.Conn, masterURL string) *RPCClient {
	conn.SetReadLimit(protocol.SocketFrameSize)
	client := jsonrpc.NewClient(conn)
	return &RPCClient{
		client:    client,
		masterURL: masterURL,
	}
}

func (rpc *RPCClient) safeCall(serviceMethod string, args interface{}, reply interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Recovered from panic in %s: %v", serviceMethod, r)
		}
	}()
	return rpc.client.Call(serviceMethod, args, reply)
}

func (rpc *RPCClient) Ping() error {
	return rpc.safeCall("RPCNode.Ping", &Void{}, &Void{})
}

func (rpc *RPCClient) GetDevices() (map[string][]device.Device, error) {
	logrus.Info("RPCNode.GetDevices")

	var resp DevicesResponse
	if err := rpc.safeCall("RPCNode.GetDevices", &Void{}, &resp); err != nil {
		return nil, err
	}

	var deviceMap map[string][]device.Device
	deviceMap = make(map[string][]device.Device)
	for _, d := range resp.Devices {
		deviceMap[d.Manager] = append(deviceMap[d.Manager], &RPCDevice{
			Dev: d,
		})
	}
	return deviceMap, nil
}

func (rpc *RPCClient) StartDevice(deviceId string) error {
	logrus.Info("RPCNode.StartDevice")

	var resp ErrorResponse
	if err := rpc.safeCall("RPCNode.StartDevice", &DeviceRequest{
		DeviceID: deviceId,
	}, &resp); err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf(resp.ErrorMessage)
	}
	return nil
}

func (rpc *RPCClient) StopDevice(deviceId string) error {
	logrus.Info("RPCNode.StopDevice")

	var resp ErrorResponse
	if err := rpc.safeCall("RPCNode.StopDevice", &DeviceRequest{
		DeviceID: deviceId,
	}, &resp); err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf(resp.ErrorMessage)
	}
	return nil
}

func getAppParameterRequest(deviceId string, parameter *app.Parameter) *AppParameterRequest {
	var appParams *AppParams
	if parameter.App != nil {
		var androidParams *AndroidParams
		if parameter.App.Android != nil {
			androidParams = &AndroidParams{
				LaunchActivity: parameter.App.Android.LaunchActivity,
			}
		}

		var executableParams *ExecutableParams
		if parameter.App.Executable != nil {
			executableParams = &ExecutableParams{
				Executable: parameter.App.Executable.Executable,
			}
		}

		appParams = &AppParams{
			AppID:      int32(parameter.App.AppBinaryID),
			Additional: parameter.App.Additional,
			Hash:       parameter.App.Hash,
			Size:       int64(parameter.App.Size),
			Android:    androidParams,
			Executable: executableParams,
		}
	}

	var webParams *WebParams
	if parameter.Web != nil {
		webParams = &WebParams{
			StartURL: parameter.Web.StartURL,
		}
	}

	return &AppParameterRequest{
		DeviceID:   deviceId,
		Platform:   parameter.Platform,
		Name:       parameter.Name,
		Identifier: parameter.Identifier,
		Version:    parameter.Version,
		App:        appParams,
		Web:        webParams,
	}
}

func (rpc *RPCClient) IsAppInstalled(deviceId string, parameter *app.Parameter) (bool, error) {
	logrus.Info("RPCNode.IsAppInstalled")

	var resp BoolResponse
	if err := rpc.safeCall("RPCNode.IsAppInstalled", getAppParameterRequest(deviceId, parameter), &resp); err != nil {
		return false, err
	}
	if resp.ErrorCode != 0 {
		return false, fmt.Errorf(resp.ErrorMessage)
	}
	return resp.Value, nil
}

func (rpc *RPCClient) IsAppUploaded(parameter *app.Parameter) (bool, error) {
	logrus.Info("RPCNode.IsAppUploaded")

	var resp BoolResponse
	if err := rpc.safeCall("RPCNode.IsAppUploaded", getAppParameterRequest("", parameter), &resp); err != nil {
		return false, err
	}
	if resp.ErrorCode != 0 {
		return false, fmt.Errorf(resp.ErrorMessage)
	}
	return resp.Value, nil
}

func (rpc *RPCClient) UploadApp(ctx context.Context, parameter *app.Parameter) error {
	logrus.Info("RPCNode.UploadApp")

	filename := filepath.Base(parameter.App.AppPath)

	fileURL, _ := url.JoinPath("http://"+rpc.masterURL, "upload", filename)
	appId := int32(parameter.App.AppBinaryID)
	appSize := int64(parameter.App.Size)
	uploadRequest := &UploadAppRequest{
		AppID:      appId,
		Identifier: parameter.Identifier,
		Name:       filename,
		Hash:       parameter.App.Hash,
		Size:       appSize,
		URL:        fileURL,
	}

	var resp BoolResponse
	if err := rpc.safeCall("RPCNode.UploadApp", uploadRequest, &resp); err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf(resp.ErrorMessage)
	}

	var progressResponse UploadAppProgressResponse
	for {
		time.Sleep(time.Second)
		if err := rpc.safeCall("RPCNode.UploadAppProgress", UploadAppProgressRequest{AppID: appId}, &progressResponse); err != nil {
			return err
		}
		if resp.ErrorCode != 0 {
			return fmt.Errorf(resp.ErrorMessage)
		}

		if progressResponse.DataReceived == appSize {
			logrus.Info("RPCNode.UploadApp complete")
			break
		}
	}
	return nil
}

func (rpc *RPCClient) InstallApp(deviceId string, parameter *app.Parameter) error {
	logrus.Info("RPCNode.InstallApp")
	var resp BoolResponse
	if err := rpc.safeCall("RPCNode.InstallApp", getAppParameterRequest(deviceId, parameter), &resp); err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf(resp.ErrorMessage)
	}
	return nil
}

func (rpc *RPCClient) UninstallApp(deviceId string, parameter *app.Parameter) error {
	logrus.Info("RPCNode.UninstallApp")
	var resp BoolResponse
	if err := rpc.safeCall("RPCNode.UninstallApp", getAppParameterRequest(deviceId, parameter), &resp); err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf(resp.ErrorMessage)
	}
	return nil
}

func (rpc *RPCClient) StartApp(deviceId string, parameter *app.Parameter, sessionId string, nodeUrl string) error {
	logrus.Info("RPCNode.StartApp")

	var resp BoolResponse
	if err := rpc.safeCall("RPCNode.StartApp", &StartAppRequest{
		App:       getAppParameterRequest(deviceId, parameter),
		SessionID: sessionId,
		HostIP:    nodeUrl,
	}, &resp); err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf(resp.ErrorMessage)
	}
	return nil
}

func (rpc *RPCClient) StopApp(deviceId string, parameter *app.Parameter) error {
	logrus.Info("RPCNode.StopApp")

	var resp BoolResponse
	if err := rpc.safeCall("RPCNode.StopApp", getAppParameterRequest(deviceId, parameter), &resp); err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf(resp.ErrorMessage)
	}
	return nil
}

func (rpc *RPCClient) IsConnected(deviceId string) bool {
	logrus.Info("RPCNode.IsConnected")

	var resp BoolResponse
	if err := rpc.safeCall("RPCNode.IsConnected", &DeviceRequest{DeviceID: deviceId}, &resp); err != nil {
		return false
	}
	return resp.Value
}

func (rpc *RPCClient) StartRecording(deviceId string, path string) error {
	return nil
}

func (rpc *RPCClient) StopRecording(deviceId string) error {
	return nil
}

func (rpc *RPCClient) GetScreenshot(deviceId string) ([]byte, int, int, error) {
	logrus.Info("RPCNode.GetScreenshot")

	var resp ScreenShotResponse
	if err := rpc.safeCall("RPCNode.TakeScreenshot", &DeviceRequest{DeviceID: deviceId}, &resp); err != nil {
		return nil, 0, 0, err
	}

	screenshotData := make([]byte, 0, resp.Size)
	chunkSize := 1024
	for start := 0; start < int(resp.Size); start += chunkSize {
		end := start + chunkSize
		if end > int(resp.Size) {
			end = int(resp.Size)
		}

		var dataResp ScreenShotDataResponse
		if err := rpc.safeCall("RPCNode.GetScreenshotData", &ScreenShotDataRequest{
			Hash:  resp.Hash,
			Start: int32(start),
			End:   int32(end),
		}, &dataResp); err != nil {
			return nil, 0, 0, err
		}

		screenshotData = append(screenshotData, dataResp.Data...)
	}

	return screenshotData, int(resp.Width), int(resp.Height), nil
}

func (rpc *RPCClient) HasFeature(deviceId string, feature string) bool {
	logrus.Info("RPCNode.HasFeature")

	var resp BoolResponse
	if err := rpc.safeCall("RPCNode.HasFeature", &FeatureRequest{DeviceID: deviceId, Feature: feature}, &resp); err != nil {
		return false
	}
	if resp.ErrorCode != 0 {
		return false
	}
	return resp.Value
}

func (rpc *RPCClient) Execute(deviceId string, data string) {
	logrus.Info("RPCNode.Execute")

	_ = rpc.safeCall("RPCNode.Execute", &ExecuteRequest{DeviceID: deviceId, Data: data}, &Void{})
}

func (rpc *RPCClient) ConnectionTimeout(deviceId string) time.Duration {
	logrus.Info("RPCNode.ConnectionTimeout")

	var resp TimeoutResponse
	if err := rpc.safeCall("RPCNode.ConnectionTimeout", &DeviceRequest{DeviceID: deviceId}, &resp); err != nil {
		return 5 * time.Minute
	}
	return time.Duration(resp.Timeout)
}

func (rpc *RPCClient) RunNativeScript(deviceId string, script []byte) {
	logrus.Info("RPCNode.RunNativeScript")
	_ = rpc.safeCall("RPCNode.RunNativeScript", &ExecuteRequest{DeviceID: deviceId, Data: string(script)}, &Void{})
}

func (rpc *RPCClient) SendAction(deviceId string, script []byte) {
	logrus.Info("RPCNode.SendAction")
	_ = rpc.safeCall("RPCNode.SendAction", &ExecuteRequest{DeviceID: deviceId, Data: string(script)}, &Void{})
}
