package mcp

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/device"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultHierarchyChars = 60000
	defaultLogLines       = 200
	maxLogLines           = 2000
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

type LogsArgs struct {
	DeviceID uint `json:"deviceId" jsonschema:"device id from list_devices"`
	Lines    int  `json:"lines,omitempty" jsonschema:"number of most recent lines (default 200, max 2000)"`
}

// takeScreenshot asks the device, or with source "app" its connected app, for
// a screenshot and returns it with the size of the device screen.
func (t *tools) takeScreenshot(ctx context.Context, dev device.Device, source string) ([]byte, int, int, error) {
	if source == "app" {
		return t.Control.AppScreenshot(ctx, dev)
	}
	return t.Control.Screenshot(dev)
}

func (t *tools) screenshot(ctx context.Context, req *mcp.CallToolRequest, in ScreenshotArgs) (*mcp.CallToolResult, any, error) {
	dev, err := t.liveDevice(ctx, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	data, width, height, err := t.takeScreenshot(ctx, dev, in.Source)
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
	scaled, jpg, err := scaleToJPEG(img, in.MaxWidth)
	if err != nil {
		return nil, nil, err
	}
	scale := float64(width) / float64(scaled.Bounds().Dx())
	return jpegResult(jpg, fmt.Sprintf("device screen %dx%d px; image %dx%d (multiply image coordinates by %.3f for tap/swipe)",
		width, height, scaled.Bounds().Dx(), scaled.Bounds().Dy(), scale))
}

// selectElements keeps the elements of the XML that match the XPath.
func selectElements(xml []byte, xpath string) (string, error) {
	doc, err := xmlquery.Parse(bytes.NewReader(xml))
	if err != nil {
		return "", err
	}
	nodes, err := xmlquery.QueryAll(doc, xpath)
	if err != nil {
		return "", fmt.Errorf("invalid xpath: %w", err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d matching elements\n", len(nodes))
	for _, n := range nodes {
		b.WriteString(n.OutputXML(true))
		b.WriteString("\n")
	}
	return b.String(), nil
}

// truncate cuts content to maxChars (default defaultHierarchyChars) and says
// so.
func truncate(content string, maxChars int) string {
	if maxChars <= 0 {
		maxChars = defaultHierarchyChars
	}
	if len(content) <= maxChars {
		return content
	}
	return content[:maxChars] + fmt.Sprintf("\n... truncated (%d characters total); use xpath to narrow it down", len(content))
}

func (t *tools) getHierarchy(ctx context.Context, req *mcp.CallToolRequest, in HierarchyArgs) (*mcp.CallToolResult, any, error) {
	dev, err := t.liveDevice(ctx, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	xml, err := t.Control.Hierarchy(ctx, dev, false)
	if err != nil {
		return nil, nil, err
	}
	content := string(xml)
	if in.XPath != "" {
		if content, err = selectElements(xml, in.XPath); err != nil {
			return nil, nil, err
		}
	}
	return textResult("%s", truncate(content, in.MaxChars))
}

func (t *tools) deviceLogs(ctx context.Context, req *mcp.CallToolRequest, in LogsArgs) (*mcp.CallToolResult, any, error) {
	dev, err := t.liveDevice(ctx, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	lines := in.Lines
	if lines <= 0 {
		lines = defaultLogLines
	}
	result, err := t.Control.Logs(dev, min(lines, maxLogLines))
	if err != nil {
		return nil, nil, err
	}
	return textResult("%s", strings.Join(result, "\n"))
}
