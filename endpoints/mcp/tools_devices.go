package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DeviceArgs struct {
	DeviceID uint `json:"deviceId" jsonschema:"device id from list_devices"`
}

type ListDevicesArgs struct {
	Platform string `json:"platform,omitempty" jsonschema:"optional filter: android, ios, ios_simulator, mac, windows, linux, web, unity_editor"`
}

type deviceInfo struct {
	ID           uint              `json:"id"`
	Identifier   string            `json:"identifier"`
	Name         string            `json:"name"`
	Alias        string            `json:"alias,omitempty"`
	Platform     string            `json:"platform"`
	OS           string            `json:"os"`
	OSVersion    string            `json:"osVersion"`
	Node         string            `json:"node,omitempty"`
	Online       bool              `json:"online"`
	Status       string            `json:"status"`
	Locked       bool              `json:"locked"`
	LockedBy     string            `json:"lockedBy,omitempty"`
	AppConnected bool              `json:"appConnected"`
	Parameter    map[string]string `json:"parameter,omitempty"`
}

func (t *tools) registerDevices(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "list_devices", Description: "List the devices of the project with status, lock holder and whether an app is connected.", Annotations: readOnly()}, t.listDevices)
	mcp.AddTool(s, &mcp.Tool{Name: "get_device", Description: "Get one device with its parameters.", Annotations: readOnly()}, t.getDevice)
	mcp.AddTool(s, &mcp.Tool{Name: "unlock_device", Description: "Force unlock a device. This aborts a test run or control session that holds it.", Annotations: destructive()}, t.unlockDevice)
}

// deviceInfo describes a stored device together with what the pool and the
// leases know of it right now.
func (t *tools) deviceInfo(d *models.Device, details bool) deviceInfo {
	info := deviceInfo{
		ID: d.ID, Identifier: d.DeviceIdentifier, Name: d.Name, Alias: d.Alias,
		Platform: platformNames()[d.PlatformType], OS: d.OS, OSVersion: d.OSVersion,
		Status: device.StateToString(device.StateUnknown),
	}
	if d.Node != nil {
		info.Node = d.Node.Name
	}
	dev, _ := t.Pool.GetDevice(d.DeviceIdentifier)
	switch {
	case dev != nil:
		t.addLiveState(&info, dev, details)
	case d.NodeID > 0:
		info.Status = device.StateToString(device.StateNodeDisconnected)
	}
	return info
}

func (t *tools) addLiveState(info *deviceInfo, dev device.Device, details bool) {
	info.Online = true
	info.Status = device.StateToString(dev.DeviceState())
	info.Locked = dev.IsLocked()
	info.AppConnected = dev.IsAppConnected()
	if l, ok := t.Control.Holder(dev.DeviceID()); ok {
		info.LockedBy = l.Owner
	}
	if details {
		info.Parameter = dev.DeviceParameter()
	}
}

// platformFilter validates the platform argument of list_devices; "" lists
// every platform.
func platformFilter(arg string) (string, error) {
	platform := strings.ToLower(arg)
	if platform == "" {
		return "", nil
	}
	for _, name := range platformNames() {
		if name == platform {
			return platform, nil
		}
	}
	return "", fmt.Errorf("unknown platform %q", arg)
}

func (t *tools) listDevices(ctx context.Context, req *mcp.CallToolRequest, in ListDevicesArgs) (*mcp.CallToolResult, any, error) {
	platform, err := platformFilter(in.Platform)
	if err != nil {
		return nil, nil, err
	}
	devices, err := t.Lookup.Devices(ctx, t.project)
	if err != nil {
		return nil, nil, err
	}
	result := make([]deviceInfo, 0, len(devices))
	for i := range devices {
		if platform == "" || platformNames()[devices[i].PlatformType] == platform {
			result = append(result, t.deviceInfo(&devices[i], false))
		}
	}
	return jsonResult(result)
}

func (t *tools) getDevice(ctx context.Context, req *mcp.CallToolRequest, in DeviceArgs) (*mcp.CallToolResult, any, error) {
	d, err := t.Lookup.Device(ctx, t.project, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(t.deviceInfo(d, true))
}

func (t *tools) unlockDevice(ctx context.Context, req *mcp.CallToolRequest, in DeviceArgs) (*mcp.CallToolResult, any, error) {
	if _, err := t.Devices.Unlock(ctx, t.project, strconv.FormatUint(uint64(in.DeviceID), 10)); err != nil {
		return nil, nil, err
	}
	return textResult("device %d unlocked", in.DeviceID)
}
