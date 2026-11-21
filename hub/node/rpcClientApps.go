package node

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

func (c *RPCClient) IsAppInstalled(deviceId string, parameter *app.Parameter) (bool, error) {
	return c.forwardBool("IsAppInstalled", getAppParameterRequest(deviceId, parameter))
}

func (c *RPCClient) IsAppUploaded(parameter *app.Parameter) (bool, error) {
	return c.forwardBool("IsAppUploaded", getAppParameterRequest("", parameter))
}

func (c *RPCClient) InstallApp(deviceId string, parameter *app.Parameter) error {
	return c.forwardOK("InstallApp", getAppParameterRequest(deviceId, parameter))
}

func (c *RPCClient) UninstallApp(deviceId string, parameter *app.Parameter) error {
	return c.forwardOK("UninstallApp", getAppParameterRequest(deviceId, parameter))
}

//nolint:revive // argument-limit: fixed by manager.RPCApps
func (c *RPCClient) StartApp(deviceId string, config *device.DeviceConfig, parameter *app.Parameter, sessionId string, nodeUrl string) error {
	req := startAppRequest(getAppParameterRequest(deviceId, parameter), config)
	req.SessionID, req.HostIP = sessionId, nodeUrl
	return c.forwardOK("StartApp", req)
}

func (c *RPCClient) StopApp(deviceId string, parameter *app.Parameter) error {
	return c.forwardOK("StopApp", getAppParameterRequest(deviceId, parameter))
}

func (c *RPCClient) ClearAppData(deviceId string, parameter *app.Parameter) error {
	return c.forwardOK("ClearAppData", getAppParameterRequest(deviceId, parameter))
}
