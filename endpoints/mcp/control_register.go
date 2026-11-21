package mcp

import "github.com/modelcontextprotocol/go-sdk/mcp"

func (t *tools) registerControl(s *mcp.Server) {
	t.registerLeaseTools(s)
	t.registerInputTools(s)
	t.registerAppTools(s)
}

func (t *tools) registerLeaseTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "lock_device", Description: "Take exclusive control of a device. Required before tap, swipe, type_text, press_key and the app tools. Fails while a test run uses the device. The lock expires after 10 idle minutes and when the MCP session ends.", Annotations: writeSafe()}, t.lockDevice)
	mcp.AddTool(s, &mcp.Tool{Name: "release_device", Description: "Give up control of a device locked with lock_device.", Annotations: writeSafe()}, t.releaseDevice)
	mcp.AddTool(s, &mcp.Tool{Name: "screenshot", Description: "Take a screenshot of a device. The image may be downscaled; tap and swipe coordinates are in the original device pixels reported with it.", Annotations: readOnly()}, t.screenshot)
	mcp.AddTool(s, &mcp.Tool{Name: "get_hierarchy", Description: "Scene graph of the connected Unity app as XML (elements with ID, Name, Class, position, isVisible). Optionally filtered by XPath.", Annotations: readOnly()}, t.getHierarchy)
	mcp.AddTool(s, &mcp.Tool{Name: "device_logs", Description: "Most recent system log lines of the device (Android logcat).", Annotations: readOnly()}, t.deviceLogs)
}

func (t *tools) registerInputTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "tap", Description: "Tap a point (x/y in device pixels) or touch an app element by xpath or elementId. Needs lock_device.", Annotations: writeSafe()}, t.tap)
	mcp.AddTool(s, &mcp.Tool{Name: "swipe", Description: "Swipe between two points in device pixels. Needs lock_device.", Annotations: writeSafe()}, t.swipe)
	mcp.AddTool(s, &mcp.Tool{Name: "type_text", Description: "Type text into the focused input. Needs lock_device.", Annotations: writeSafe()}, t.typeText)
	mcp.AddTool(s, &mcp.Tool{Name: "press_key", Description: "Press a hardware or system key. Needs lock_device.", Annotations: writeSafe()}, t.pressKey)
}

func (t *tools) registerAppTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "install_app", Description: "Install an app bundle on the device. Needs lock_device.", Annotations: writeSafe()}, t.installApp)
	mcp.AddTool(s, &mcp.Tool{Name: "start_app", Description: "Start an app bundle on the device; the app connects back so get_hierarchy and element taps work. Needs lock_device.", Annotations: writeSafe()}, t.startApp)
	mcp.AddTool(s, &mcp.Tool{Name: "stop_app", Description: "Stop the app. With clearData the app data is wiped. Needs lock_device.", Annotations: destructive()}, t.stopApp)
}
