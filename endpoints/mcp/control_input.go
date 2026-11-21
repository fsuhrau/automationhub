package mcp

import (
	"context"
	"fmt"
	"time"

	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const defaultSwipe = 300 * time.Millisecond

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

func (t *tools) tap(ctx context.Context, req *mcp.CallToolRequest, in TapArgs) (*mcp.CallToolResult, any, error) {
	target, err := t.leased(ctx, req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	ref := hubcontrol.ElementRef{ID: in.ElementID, XPath: in.XPath}
	switch {
	case !ref.IsZero():
		if err := t.Control.TouchElement(ctx, target, ref); err != nil {
			return nil, nil, err
		}
		return textResult("touched element %s%s", in.ElementID, in.XPath)
	case in.X != nil && in.Y != nil:
		if err := t.Control.Tap(target, *in.X, *in.Y); err != nil {
			return nil, nil, err
		}
		return textResult("tapped %d,%d", *in.X, *in.Y)
	}
	return nil, nil, fmt.Errorf("give x and y, or xpath or elementId")
}

func (t *tools) swipe(ctx context.Context, req *mcp.CallToolRequest, in SwipeArgs) (*mcp.CallToolResult, any, error) {
	target, err := t.leased(ctx, req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	d := time.Duration(in.DurationMs) * time.Millisecond
	if d <= 0 {
		d = defaultSwipe
	}
	gesture := hubcontrol.SwipeGesture{FromX: in.FromX, FromY: in.FromY, ToX: in.ToX, ToY: in.ToY, Duration: d}
	if err := t.Control.Swipe(target, gesture); err != nil {
		return nil, nil, err
	}
	return textResult("swiped %d,%d -> %d,%d", in.FromX, in.FromY, in.ToX, in.ToY)
}

func (t *tools) typeText(ctx context.Context, req *mcp.CallToolRequest, in TextArgs) (*mcp.CallToolResult, any, error) {
	target, err := t.leased(ctx, req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	if err := t.Control.InputText(target, in.Text); err != nil {
		return nil, nil, err
	}
	return textResult("typed %d characters", len([]rune(in.Text)))
}

func (t *tools) pressKey(ctx context.Context, req *mcp.CallToolRequest, in KeyArgs) (*mcp.CallToolResult, any, error) {
	target, err := t.leased(ctx, req, in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	if err := t.Control.PressKey(target, in.Key); err != nil {
		return nil, nil, err
	}
	return textResult("pressed %s", in.Key)
}
