package mcp

import (
	"context"
	"fmt"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"

	"github.com/fsuhrau/automationhub/device"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// liveDevice resolves the connected device for a project device id.
func (t *tools) liveDevice(ctx context.Context, id uint) (device.Device, error) {
	d, err := t.Lookup.Device(ctx, t.project, id)
	if err != nil {
		return nil, err
	}
	dev, _ := t.Pool.GetDevice(d.DeviceIdentifier)
	if dev == nil {
		return nil, fmt.Errorf("device %d (%s) is not connected", id, d.Name)
	}
	return dev, nil
}

// leased resolves the device and the lease token of this session.
func (t *tools) leased(ctx context.Context, req *mcp.CallToolRequest, id uint) (hubcontrol.Target, error) {
	dev, err := t.liveDevice(ctx, id)
	if err != nil {
		return hubcontrol.Target{}, err
	}
	token, err := t.Control.TokenFor(dev.DeviceID(), t.ownerFor(req.Session))
	if err != nil {
		return hubcontrol.Target{}, fmt.Errorf("%w (call lock_device with deviceId %d first)", err, id)
	}
	return hubcontrol.Target{Device: dev, Token: token}, nil
}

func (t *tools) lockDevice(ctx context.Context, req *mcp.CallToolRequest, in DeviceArgs) (*mcp.CallToolResult, any, error) {
	dev, err := t.liveDevice(ctx, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	lease, err := t.Control.Acquire(dev, t.ownerFor(req.Session))
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{"deviceId": in.DeviceID, "owner": lease.Owner, "expiresAt": lease.ExpiresAt, "appConnected": dev.IsAppConnected()})
}

func (t *tools) releaseDevice(ctx context.Context, req *mcp.CallToolRequest, in DeviceArgs) (*mcp.CallToolResult, any, error) {
	target, err := t.leased(ctx, req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	if err := t.Control.Release(target.Device.DeviceID(), target.Token); err != nil {
		return nil, nil, err
	}
	return textResult("device %d released", in.DeviceID)
}
