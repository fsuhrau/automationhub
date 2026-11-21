package iosdevice

import (
	"fmt"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/luascript"
	"github.com/fsuhrau/automationhub/modules/webdriver"
)

// wdaClient is what the WebDriver strategy needs from the WebDriverAgent client.
type wdaClient interface {
	FindElement(by, value string) (*webdriver.Element, error)
	TapElement(elementID string) error
	SendText(text string) error
	GetAlertText() (string, error)
}

// LuaScriptHandler runs hub scripts on an iOS device.
type LuaScriptHandler struct {
	*luascript.Handler
}

// New creates the script handler of an iOS device.
func New(client *webdriver.Client, dev device.Device) *LuaScriptHandler {
	return newScriptHandler(client, dev, luascript.SystemClock())
}

func newScriptHandler(client wdaClient, dev luascript.Logger, clock luascript.Clock) *LuaScriptHandler {
	wda := &webDriverScreen{client: client}
	return &LuaScriptHandler{luascript.New(luascript.Deps{
		Finder: wda, Clicker: wda, Keyboard: wda, Alerts: wda, Logger: dev, Clock: clock,
	})}
}

// webDriverScreen finds and taps elements through WebDriverAgent.
type webDriverScreen struct {
	client wdaClient
}

// Find reports whether WebDriverAgent knows an element for the lookup.
func (w *webDriverScreen) Find(lookupType, lookup string) (bool, error) {
	element, err := w.client.FindElement(lookupType, lookup)
	if err != nil {
		return false, fmt.Errorf("find element failed: %w", err)
	}
	return element != nil, nil
}

// Click taps the element; every failure is raised in the script.
func (w *webDriverScreen) Click(lookupType, lookup string) error {
	element, err := w.client.FindElement(lookupType, lookup)
	if err != nil || element == nil || element.Value == nil || len(element.Value.Element) == 0 {
		return fmt.Errorf("tap failed unable to find element: %s", lookup)
	}
	if err := w.client.TapElement(element.Value.Element); err != nil {
		return fmt.Errorf("tap failed: %s", lookup)
	}
	return nil
}

// SendText types the text through WebDriverAgent.
func (w *webDriverScreen) SendText(text string) error {
	return w.client.SendText(text)
}

// AlertText returns the text of the visible alert.
func (w *webDriverScreen) AlertText() (string, error) {
	return w.client.GetAlertText()
}
