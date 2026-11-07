package mcp

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/image/draw"
)

const (
	defaultScreenshotWidth = 1024
	defaultHierarchyChars  = 60000
)

type ScreenshotArgs struct {
	DeviceID uint   `json:"deviceId" jsonschema:"device id from list_devices"`
	MaxWidth int    `json:"maxWidth,omitempty" jsonschema:"downscale to at most this width (default 1024, 0 keeps the default)"`
	Source   string `json:"source,omitempty" jsonschema:"native (default) or app: ask the connected Unity app instead of the OS"`
}

type HierarchyArgs struct {
	DeviceID uint   `json:"deviceId" jsonschema:"device id from list_devices"`
	XPath    string `json:"xpath,omitempty" jsonschema:"only return elements matching this XPath, e.g. //Button[@Name='Play']"`
	MaxChars int    `json:"maxChars,omitempty" jsonschema:"truncate the XML after this many characters (default 60000)"`
}

type TapArgs struct {
	DeviceID  uint   `json:"deviceId" jsonschema:"device id from list_devices"`
	X         *int   `json:"x,omitempty" jsonschema:"x in device screenshot pixels"`
	Y         *int   `json:"y,omitempty" jsonschema:"y in device screenshot pixels"`
	XPath     string `json:"xpath,omitempty" jsonschema:"touch the app element matching this XPath instead of a point"`
	ElementID string `json:"elementId,omitempty" jsonschema:"touch the app element with this ID instead of a point"`
}

type SwipeArgs struct {
	DeviceID   uint `json:"deviceId" jsonschema:"device id from list_devices"`
	FromX      int  `json:"fromX" jsonschema:"start x in device screenshot pixels"`
	FromY      int  `json:"fromY" jsonschema:"start y in device screenshot pixels"`
	ToX        int  `json:"toX" jsonschema:"end x in device screenshot pixels"`
	ToY        int  `json:"toY" jsonschema:"end y in device screenshot pixels"`
	DurationMs int  `json:"durationMs,omitempty" jsonschema:"swipe duration in milliseconds (default 300)"`
}

type TextArgs struct {
	DeviceID uint   `json:"deviceId" jsonschema:"device id from list_devices"`
	Text     string `json:"text" jsonschema:"text to type into the focused field"`
}

type KeyArgs struct {
	DeviceID uint   `json:"deviceId" jsonschema:"device id from list_devices"`
	Key      string `json:"key" jsonschema:"home, back, enter, menu, delete, tab, app_switch, power, volume_up, volume_down (iOS: home and volume keys only)"`
}

type AppBundleArgs struct {
	DeviceID uint `json:"deviceId" jsonschema:"device id from list_devices"`
	BundleID uint `json:"bundleId" jsonschema:"app bundle id from list_bundles"`
}

type StopAppArgs struct {
	DeviceID  uint `json:"deviceId" jsonschema:"device id from list_devices"`
	BundleID  uint `json:"bundleId" jsonschema:"app bundle id from list_bundles"`
	ClearData bool `json:"clearData,omitempty" jsonschema:"also wipe the app data (irreversible)"`
}

type LogsArgs struct {
	DeviceID uint `json:"deviceId" jsonschema:"device id from list_devices"`
	Lines    int  `json:"lines,omitempty" jsonschema:"number of most recent lines (default 200, max 2000)"`
}

func (t *tools) registerControl(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "lock_device", Description: "Take exclusive control of a device. Required before tap, swipe, type_text, press_key and the app tools. Fails while a test run uses the device. The lock expires after 10 idle minutes and when the MCP session ends.", Annotations: writeSafe}, t.lockDevice)
	mcp.AddTool(s, &mcp.Tool{Name: "release_device", Description: "Give up control of a device locked with lock_device.", Annotations: writeSafe}, t.releaseDevice)
	mcp.AddTool(s, &mcp.Tool{Name: "screenshot", Description: "Take a screenshot of a device. The image may be downscaled; tap and swipe coordinates are in the original device pixels reported with it.", Annotations: readOnly}, t.screenshot)
	mcp.AddTool(s, &mcp.Tool{Name: "get_hierarchy", Description: "Scene graph of the connected Unity app as XML (elements with ID, Name, Class, position, isVisible). Optionally filtered by XPath.", Annotations: readOnly}, t.getHierarchy)
	mcp.AddTool(s, &mcp.Tool{Name: "tap", Description: "Tap a point (x/y in device pixels) or touch an app element by xpath or elementId. Needs lock_device.", Annotations: writeSafe}, t.tap)
	mcp.AddTool(s, &mcp.Tool{Name: "swipe", Description: "Swipe between two points in device pixels. Needs lock_device.", Annotations: writeSafe}, t.swipe)
	mcp.AddTool(s, &mcp.Tool{Name: "type_text", Description: "Type text into the focused input. Needs lock_device.", Annotations: writeSafe}, t.typeText)
	mcp.AddTool(s, &mcp.Tool{Name: "press_key", Description: "Press a hardware or system key. Needs lock_device.", Annotations: writeSafe}, t.pressKey)
	mcp.AddTool(s, &mcp.Tool{Name: "install_app", Description: "Install an app bundle on the device. Needs lock_device.", Annotations: writeSafe}, t.installApp)
	mcp.AddTool(s, &mcp.Tool{Name: "start_app", Description: "Start an app bundle on the device; the app connects back so get_hierarchy and element taps work. Needs lock_device.", Annotations: writeSafe}, t.startApp)
	mcp.AddTool(s, &mcp.Tool{Name: "stop_app", Description: "Stop the app. With clearData the app data is wiped. Needs lock_device.", Annotations: destructive}, t.stopApp)
	mcp.AddTool(s, &mcp.Tool{Name: "device_logs", Description: "Most recent system log lines of the device (Android logcat).", Annotations: readOnly}, t.deviceLogs)
}

// liveDevice resolves the connected device for a project device id.
func (t *tools) liveDevice(id uint) (device.Device, error) {
	d, err := t.findDevice(id)
	if err != nil {
		return nil, err
	}
	dev, _ := t.Devices.GetDevice(d.DeviceIdentifier)
	if dev == nil {
		return nil, fmt.Errorf("device %d (%s) is not connected", id, d.Name)
	}
	return dev, nil
}

// leased resolves the device and the lease token of this session.
func (t *tools) leased(req *mcp.CallToolRequest, id uint) (device.Device, string, error) {
	dev, err := t.liveDevice(id)
	if err != nil {
		return nil, "", err
	}
	token, err := t.Control.TokenFor(dev.DeviceID(), t.ownerFor(req.Session))
	if err != nil {
		return nil, "", fmt.Errorf("%w (call lock_device with deviceId %d first)", err, id)
	}
	return dev, token, nil
}

func (t *tools) lockDevice(ctx context.Context, req *mcp.CallToolRequest, in DeviceArgs) (*mcp.CallToolResult, any, error) {
	dev, err := t.liveDevice(in.DeviceID)
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
	dev, token, err := t.leased(req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	if err := t.Control.Release(dev.DeviceID(), token); err != nil {
		return nil, nil, err
	}
	return textResult("device %d released", in.DeviceID)
}

func (t *tools) screenshot(ctx context.Context, req *mcp.CallToolRequest, in ScreenshotArgs) (*mcp.CallToolResult, any, error) {
	dev, err := t.liveDevice(in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	var data []byte
	var width, height int
	if in.Source == "app" {
		data, width, height, err = t.Control.AppScreenshot(ctx, dev)
	} else {
		data, width, height, err = t.Control.Screenshot(dev)
	}
	if err != nil {
		return nil, nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("decode screenshot: %w", err)
	}
	if width == 0 || height == 0 {
		width, height = img.Bounds().Dx(), img.Bounds().Dy()
	}
	maxWidth := in.MaxWidth
	if maxWidth <= 0 {
		maxWidth = defaultScreenshotWidth
	}
	scaled := img
	if b := img.Bounds(); b.Dx() > maxWidth {
		h := max(1, (b.Dy()*maxWidth+b.Dx()/2)/b.Dx())
		dst := image.NewRGBA(image.Rect(0, 0, maxWidth, h))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Src, nil)
		scaled = dst
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, scaled, &jpeg.Options{Quality: 80}); err != nil {
		return nil, nil, err
	}
	scale := float64(width) / float64(scaled.Bounds().Dx())
	note := fmt.Sprintf("device screen %dx%d px; image %dx%d (multiply image coordinates by %.3f for tap/swipe)",
		width, height, scaled.Bounds().Dx(), scaled.Bounds().Dy(), scale)
	return &mcp.CallToolResult{Content: []mcp.Content{
		&mcp.ImageContent{Data: buf.Bytes(), MIMEType: "image/jpeg"},
		&mcp.TextContent{Text: note},
	}}, nil, nil
}

func (t *tools) getHierarchy(ctx context.Context, req *mcp.CallToolRequest, in HierarchyArgs) (*mcp.CallToolResult, any, error) {
	dev, err := t.liveDevice(in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	xml, err := t.Control.Hierarchy(ctx, dev, false)
	if err != nil {
		return nil, nil, err
	}
	content := string(xml)
	if in.XPath != "" {
		doc, err := xmlquery.Parse(bytes.NewReader(xml))
		if err != nil {
			return nil, nil, err
		}
		nodes, err := xmlquery.QueryAll(doc, in.XPath)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid xpath: %w", err)
		}
		var b strings.Builder
		fmt.Fprintf(&b, "%d matching elements\n", len(nodes))
		for _, n := range nodes {
			b.WriteString(n.OutputXML(true))
			b.WriteString("\n")
		}
		content = b.String()
	}
	maxChars := in.MaxChars
	if maxChars <= 0 {
		maxChars = defaultHierarchyChars
	}
	if len(content) > maxChars {
		content = content[:maxChars] + fmt.Sprintf("\n... truncated (%d characters total); use xpath to narrow it down", len(content))
	}
	return textResult("%s", content)
}

func (t *tools) tap(ctx context.Context, req *mcp.CallToolRequest, in TapArgs) (*mcp.CallToolResult, any, error) {
	dev, token, err := t.leased(req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	ref := hubcontrol.ElementRef{ID: in.ElementID, XPath: in.XPath}
	switch {
	case !ref.IsZero():
		if err := t.Control.TouchElement(ctx, dev, token, ref); err != nil {
			return nil, nil, err
		}
		return textResult("touched element %s%s", in.ElementID, in.XPath)
	case in.X != nil && in.Y != nil:
		if err := t.Control.Tap(dev, token, *in.X, *in.Y); err != nil {
			return nil, nil, err
		}
		return textResult("tapped %d,%d", *in.X, *in.Y)
	}
	return nil, nil, fmt.Errorf("give x and y, or xpath or elementId")
}

func (t *tools) swipe(ctx context.Context, req *mcp.CallToolRequest, in SwipeArgs) (*mcp.CallToolResult, any, error) {
	dev, token, err := t.leased(req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	d := time.Duration(in.DurationMs) * time.Millisecond
	if d <= 0 {
		d = 300 * time.Millisecond
	}
	if err := t.Control.Swipe(dev, token, in.FromX, in.FromY, in.ToX, in.ToY, d); err != nil {
		return nil, nil, err
	}
	return textResult("swiped %d,%d -> %d,%d", in.FromX, in.FromY, in.ToX, in.ToY)
}

func (t *tools) typeText(ctx context.Context, req *mcp.CallToolRequest, in TextArgs) (*mcp.CallToolResult, any, error) {
	dev, token, err := t.leased(req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	if err := t.Control.InputText(dev, token, in.Text); err != nil {
		return nil, nil, err
	}
	return textResult("typed %d characters", len([]rune(in.Text)))
}

func (t *tools) pressKey(ctx context.Context, req *mcp.CallToolRequest, in KeyArgs) (*mcp.CallToolResult, any, error) {
	dev, token, err := t.leased(req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	if err := t.Control.PressKey(dev, token, in.Key); err != nil {
		return nil, nil, err
	}
	return textResult("pressed %s", in.Key)
}

func (t *tools) bundleParams(bundleID uint) (*app.Parameter, error) {
	var binary models.AppBinary
	err := t.DB.Preload("App").
		Joins("JOIN apps ON apps.id = app_binaries.app_id").
		Where("apps.project_id = ?", t.project.ID).
		First(&binary, "app_binaries.id = ?", bundleID).Error
	if err != nil {
		return nil, fmt.Errorf("%w: app bundle %d", errNotFound, bundleID)
	}
	params := base.GetParams(&binary, "")
	return &params, nil
}

func (t *tools) installApp(ctx context.Context, req *mcp.CallToolRequest, in AppBundleArgs) (*mcp.CallToolResult, any, error) {
	dev, token, err := t.leased(req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	params, err := t.bundleParams(in.BundleID)
	if err != nil {
		return nil, nil, err
	}
	if err := t.Control.InstallApp(ctx, dev, token, params); err != nil {
		return nil, nil, err
	}
	return textResult("installed %s %s", params.Identifier, params.Version)
}

func (t *tools) startApp(ctx context.Context, req *mcp.CallToolRequest, in AppBundleArgs) (*mcp.CallToolResult, any, error) {
	dev, token, err := t.leased(req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	params, err := t.bundleParams(in.BundleID)
	if err != nil {
		return nil, nil, err
	}
	if err := t.Control.StartApp(dev, token, params, uuid.NewString(), t.NodeURL); err != nil {
		return nil, nil, err
	}
	return textResult("started %s; it may take a few seconds until the app is connected (see get_device appConnected)", params.Identifier)
}

func (t *tools) stopApp(ctx context.Context, req *mcp.CallToolRequest, in StopAppArgs) (*mcp.CallToolResult, any, error) {
	dev, token, err := t.leased(req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	params, err := t.bundleParams(in.BundleID)
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

func (t *tools) deviceLogs(ctx context.Context, req *mcp.CallToolRequest, in LogsArgs) (*mcp.CallToolResult, any, error) {
	dev, err := t.liveDevice(in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	lines := in.Lines
	if lines <= 0 {
		lines = 200
	}
	if lines > 2000 {
		lines = 2000
	}
	result, err := t.Control.Logs(dev, lines)
	if err != nil {
		return nil, nil, err
	}
	return textResult("%s", strings.Join(result, "\n"))
}
