package mcp

import (
	"context"

	"github.com/fsuhrau/automationhub/app"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AppBundleArgs struct {
	DeviceID uint `json:"deviceId" jsonschema:"device id from list_devices"`
	BundleID uint `json:"bundleId" jsonschema:"app bundle id from list_bundles"`
}

type StopAppArgs struct {
	DeviceID  uint `json:"deviceId" jsonschema:"device id from list_devices"`
	BundleID  uint `json:"bundleId" jsonschema:"app bundle id from list_bundles"`
	ClearData bool `json:"clearData,omitempty" jsonschema:"also wipe the app data (irreversible)"`
}

// bundleParams is the install description of a bundle of the project; an
// archived bundle has no file to install.
func (t *tools) bundleParams(ctx context.Context, bundleID uint) (*app.Parameter, error) {
	return t.Devices.ControlApp(ctx, t.project, bundleID)
}

func (t *tools) installApp(ctx context.Context, req *mcp.CallToolRequest, in AppBundleArgs) (*mcp.CallToolResult, any, error) {
	dev, token, err := t.leased(ctx, req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	params, err := t.bundleParams(ctx, in.BundleID)
	if err != nil {
		return nil, nil, err
	}
	if err := t.Control.InstallApp(ctx, dev, token, params); err != nil {
		return nil, nil, err
	}
	return textResult("installed %s %s", params.Identifier, params.Version)
}

func (t *tools) startApp(ctx context.Context, req *mcp.CallToolRequest, in AppBundleArgs) (*mcp.CallToolResult, any, error) {
	dev, token, err := t.leased(ctx, req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	params, err := t.bundleParams(ctx, in.BundleID)
	if err != nil {
		return nil, nil, err
	}
	if err := t.Control.StartApp(dev, token, params, uuid.NewString(), t.NodeURL); err != nil {
		return nil, nil, err
	}
	return textResult("started %s; it may take a few seconds until the app is connected (see get_device appConnected)", params.Identifier)
}

func (t *tools) stopApp(ctx context.Context, req *mcp.CallToolRequest, in StopAppArgs) (*mcp.CallToolResult, any, error) {
	dev, token, err := t.leased(ctx, req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	params, err := t.bundleParams(ctx, in.BundleID)
	if err != nil {
		return nil, nil, err
	}
	if err := t.Control.StopApp(dev, token, params, in.ClearData); err != nil {
		return nil, nil, err
	}
	if in.ClearData {
		return textResult("stopped %s and cleared its data", params.Identifier)
	}
	return textResult("stopped %s", params.Identifier)
}
