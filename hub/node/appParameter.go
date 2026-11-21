package node

import (
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

// getAppParameterRequest is the master side: it turns an app parameter into
// the request of a device call.
func getAppParameterRequest(deviceId string, parameter *app.Parameter) *AppParameterRequest {
	return &AppParameterRequest{
		DeviceID:   deviceId,
		Platform:   parameter.Platform,
		Name:       parameter.Name,
		Identifier: parameter.Identifier,
		Version:    parameter.Version,
		App:        appParamsRequest(parameter.App),
		Web:        webParamsRequest(parameter.Web),
	}
}

func appParamsRequest(a *app.AppParams) *AppParams {
	if a == nil {
		return nil
	}
	params := &AppParams{
		AppID:      int32(a.AppBinaryID),
		Additional: a.Additional,
		Hash:       a.Hash,
		Size:       int64(a.Size),
	}
	if a.Android != nil {
		params.Android = &AndroidParams{LaunchActivity: a.Android.LaunchActivity}
	}
	if a.Executable != nil {
		params.Executable = &ExecutableParams{Executable: a.Executable.Executable}
	}
	return params
}

func webParamsRequest(w *app.WebParams) *WebParams {
	if w == nil {
		return nil
	}
	return &WebParams{StartURL: w.StartURL}
}

// startAppRequest is the master side of a StartApp call; the caller sets the
// session and the host.
func startAppRequest(appReq *AppParameterRequest, config *device.DeviceConfig) *StartAppRequest {
	var custom []*DeviceCustomParameter
	for k, v := range config.DeviceParameter {
		custom = append(custom, &DeviceCustomParameter{Key: k, Value: v})
	}
	return &StartAppRequest{
		App: appReq,
		ConnectionParams: &DeviceConnectionParams{
			Type: DeviceConnectionType(config.Connection),
			IP:   config.IP,
			Port: int64(config.Port),
		},
		CustomParameter: custom,
	}
}

// getAppParameter is the node side of getAppParameterRequest. data is the
// stored bundle the request refers to; nil when the call needs no bundle.
func getAppParameter(req *AppParameterRequest, data *AppBundleMetaData) *app.Parameter {
	return &app.Parameter{
		Platform:   req.Platform,
		Name:       req.Name,
		Identifier: req.Identifier,
		Version:    req.Version,
		App:        appParams(req.App, data),
		Web:        webParams(req.Web),
	}
}

func appParams(a *AppParams, data *AppBundleMetaData) *app.AppParams {
	if a == nil {
		return nil
	}
	params := &app.AppParams{
		Additional: a.Additional,
		Hash:       a.Hash,
		Size:       int(a.Size),
	}
	if data != nil {
		params.AppPath = data.FilePath
	}
	if a.Android != nil {
		params.Android = &app.AndroidParams{LaunchActivity: a.Android.LaunchActivity}
	}
	if a.Executable != nil {
		params.Executable = &app.ExecutableParams{Executable: a.Executable.Executable}
	}
	return params
}

func webParams(w *WebParams) *app.WebParams {
	if w == nil {
		return nil
	}
	return &app.WebParams{StartURL: w.StartURL}
}

// getDeviceConfig is the node side of a StartApp request.
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
