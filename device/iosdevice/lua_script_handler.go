package iosdevice

import (
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/modules/webdriver"
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	"strings"
	"time"
)

type LuaScriptHandler struct {
	webDriver *webdriver.Client
	dev       device.Device
}

func New(client *webdriver.Client, dev device.Device) *LuaScriptHandler {
	scriptHandler := LuaScriptHandler{
		webDriver: client,
		dev:       dev,
	}
	return &scriptHandler
}

func (l *LuaScriptHandler) Execute(script string) error {
	executor := lua.NewState()
	defer executor.Close()
	l.init(executor)
	return executor.DoString(script)
}

func (l *LuaScriptHandler) init(L *lua.LState) int {

	mod := L.RegisterModule("hub", map[string]lua.LGFunction{
		"click":            l.click,
		"waitFor":          l.waitFor,
		"exists":           l.exists,
		"waitForAlertWith": l.waitForAlertWith,
		"sendKeys":         l.sendKeys,
	})
	L.Push(mod)

	return 1
}

func (l *LuaScriptHandler) click(L *lua.LState) int {
	lookupType := L.CheckString(1)
	lookup := L.CheckString(2)

	l.dev.Log("device", "try click: %s", lookup)

	element, err := l.webDriver.FindElement(lookupType, lookup)
	if err != nil {
		L.RaiseError("tap failed unable to find element: %s", lookup)
		return 0
	}
	if element.Value == nil || len(element.Value.Element) == 0 {
		L.RaiseError("tap failed unable to find element: %s", lookup)
		return 0
	}

	if err := l.webDriver.TapElement(element.Value.Element); err != nil {
		L.RaiseError("tap failed: %s", lookup)
		return 0
	}

	return 0
}

func (l *LuaScriptHandler) waitFor(L *lua.LState) int {
	lookupType := L.CheckString(1)
	lookup := L.CheckString(2)
	duration := L.CheckInt(3)

	l.dev.Log("device", "wait for: %s duration: %d", lookup, duration)
	if err := l.waitForFunc(lookupType, lookup, duration); err != nil {
		L.RaiseError("Timeout: Element '%s' not found", lookup)
	}
	return 0
}

func (l *LuaScriptHandler) waitForFunc(lookupType, lookup string, duration int) error {
	timeout := time.Now().Add(time.Duration(duration) * time.Second)
	for {
		element, err := l.webDriver.FindElement(lookupType, lookup)
		if err != nil {
			logrus.Errorf("find element failed: %v", err)
		}

		if element == nil {
			if time.Now().After(timeout) {
				return fmt.Errorf("Timeout: wait for element '%s'", lookup)
			}
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	return nil
}

func (l *LuaScriptHandler) exists(L *lua.LState) int {
	lookupType := L.CheckString(1)
	lookup := L.CheckString(2)
	duration := L.CheckInt(3)
	L.Push(lua.LBool(l.waitForFunc(lookupType, lookup, duration) == nil))
	return 1
}

func (l *LuaScriptHandler) waitForAlertWith(L *lua.LState) int {
	contains := L.CheckString(1)
	duration := L.CheckInt(2)

	l.dev.Log("device", "waitForAlertWith: %s with timeout: %d", contains, duration)
	timeout := time.Now().Add(time.Duration(duration) * time.Second)
	for {
		text, err := l.webDriver.GetAlertText()
		if err != nil {
			logrus.Errorf("get alert text failed failed: %v", err)
		}

		if !strings.Contains(text, contains) {
			if time.Now().After(timeout) {
				L.RaiseError("Timeout: no alert found which contains'%s'", contains)
				return 0
			}
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	return 0
}

func (l *LuaScriptHandler) sendKeys(L *lua.LState) int {
	content := L.CheckString(1)
	l.dev.Log("device", "send_keys")
	err := l.webDriver.SendText(content)
	if err != nil {
		L.RaiseError("send_keys failed: %v", err)
		return 0
	}
	return 0
}
