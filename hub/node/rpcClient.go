package node

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/gorilla/websocket"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

type RPCClient struct {
	client *rpc.Client
}

func NewRPCClient(conn *websocket.Conn) *RPCClient {
	return &RPCClient{
		client: jsonrpc.NewClient(&WebSocketConn{Conn: conn}),
	}
}

func (rpc *RPCClient) Ping() error {
	return rpc.client.Call("RPCNode.Ping", &Void{}, &Void{})
}

func (rpc *RPCClient) GetDevices() (map[string][]device.Device, error) {

	var resp DevicesResponse

	if err := rpc.client.Call("RPCNode.GetDevices", &Void{}, &resp); err != nil {
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
	var resp ErrorResponse
	if err := rpc.client.Call("RPCNode.StartDevice", &DeviceRequest{
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
	var resp ErrorResponse
	if err := rpc.client.Call("RPCNode.StopDevice", &DeviceRequest{
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
	return &AppParameterRequest{
		DeviceID:       deviceId,
		AppID:          int32(parameter.AppBinaryID),
		Platform:       parameter.Platform,
		Name:           parameter.Name,
		Identifier:     parameter.Identifier,
		Version:        parameter.Version,
		LaunchActivity: parameter.LaunchActivity,
		Additional:     parameter.Additional,
		Hash:           parameter.Hash,
		Size:           int64(parameter.Size),
	}
}

func (rpc *RPCClient) IsAppInstalled(deviceId string, parameter *app.Parameter) (bool, error) {
	var resp BoolResponse
	if err := rpc.client.Call("RPCNode.IsAppInstalled", getAppParameterRequest(deviceId, parameter), &resp); err != nil {
		return false, err
	}
	if resp.ErrorCode != 0 {
		return false, fmt.Errorf(resp.ErrorMessage)
	}
	return resp.Value, nil
}

func (rpc *RPCClient) InstallApp(deviceId string, parameter *app.Parameter) error {
	var resp BoolResponse
	if err := rpc.client.Call("RPCNode.InstallApp", getAppParameterRequest(deviceId, parameter), &resp); err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf(resp.ErrorMessage)
	}
	return nil
}

func (rpc *RPCClient) UninstallApp(deviceId string, parameter *app.Parameter) error {
	var resp BoolResponse
	if err := rpc.client.Call("RPCNode.UninstallApp", getAppParameterRequest(deviceId, parameter), &resp); err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf(resp.ErrorMessage)
	}
	return nil
}

func (rpc *RPCClient) StartApp(deviceId string, parameter *app.Parameter, sessionId string, nodeUrl string) error {
	var resp BoolResponse
	if err := rpc.client.Call("RPCNode.StartApp", &StartAppRequest{
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
	var resp BoolResponse
	if err := rpc.client.Call("RPCNode.StopApp", getAppParameterRequest(deviceId, parameter), &resp); err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf(resp.ErrorMessage)
	}
	return nil
}

func (rpc *RPCClient) IsConnected(deviceId string) bool {
	var resp BoolResponse
	if err := rpc.client.Call("RPCNode.IsConnected", &DeviceRequest{DeviceID: deviceId}, &resp); err != nil {
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
	var resp ScreenShotResponse
	if err := rpc.client.Call("RPCNode.GetScreenshot", &DeviceRequest{DeviceID: deviceId}, &resp); err != nil {
		return nil, 0, 0, err
	}
	return resp.Data, int(resp.Width), int(resp.Height), nil
}

func (rpc *RPCClient) HasFeature(deviceId string, feature string) bool {
	var resp BoolResponse
	if err := rpc.client.Call("RPCNode.HasFeature", &FeatureRequest{DeviceID: deviceId, Feature: feature}, &resp); err != nil {
		return false
	}
	if resp.ErrorCode != 0 {
		return false
	}
	return resp.Value
}

func (rpc *RPCClient) Execute(deviceId string, data string) {
	_ = rpc.client.Call("RPCNode.Execute", &ExecuteRequest{DeviceID: deviceId, Data: data}, &Void{})
}

func (rpc *RPCClient) ConnectionTimeout(deviceId string) time.Duration {
	var resp TimeoutResponse
	if err := rpc.client.Call("RPCNode.ConnectionTimeout", &DeviceRequest{DeviceID: deviceId}, &resp); err != nil {
		return 5 * time.Minute
	}
	return time.Duration(resp.Timeout)
}

func (rpc *RPCClient) RunNativeScript(deviceId string, script []byte) {
	_ = rpc.client.Call("RPCNode.RunNativeScript", &ExecuteRequest{DeviceID: deviceId, Data: string(script)}, &Void{})
}

func (rpc *RPCClient) SendAction(deviceId string, script []byte) {
	_ = rpc.client.Call("RPCNode.SendAction", &ExecuteRequest{DeviceID: deviceId, Data: string(script)}, &Void{})
}
