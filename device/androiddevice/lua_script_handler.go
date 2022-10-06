package androiddevice

import (
	"fmt"
	"github.com/antchfx/xmlquery"
	lua "github.com/yuin/gopher-lua"
	"regexp"
	"strconv"
	"time"
)

type LuaScriptHandler struct {
	dev *Device
}

func New(dev *Device) *LuaScriptHandler {
	scriptHandler := LuaScriptHandler{
		dev: dev,
	}
	return &scriptHandler
}

func (l *LuaScriptHandler) Execute(script string) error {
	defer func() {
		if err := recover(); err != nil {
			if l.dev != nil {
				l.dev.Error("device", "script execution failed: %v", err)
			}
		}
	}()

	executor := lua.NewState()
	defer executor.Close()
	l.init(executor)
	return executor.DoString(script)
}

func (l *LuaScriptHandler) init(L *lua.LState) int {

	mod := L.RegisterModule("hub", map[string]lua.LGFunction{
		"click":   l.click,
		"waitFor": l.waitFor,
		"exists":  l.exists,
		// "waitForAlertWith": l.waitForAlertWith,
		"sendKeys": l.sendKeys,
	})
	L.Push(mod)

	return 1
}

func (l *LuaScriptHandler) click(L *lua.LState) int {
	lookupType := L.CheckString(1)
	lookup := L.CheckString(2)

	if lookupType != "xpath" {
		l.dev.Error("device", "native script lookupType '%s' not supported", lookupType)
		return 0
	}

	boundsEx := regexp.MustCompile(`\[([0-9]+),([0-9]+)\]\[([0-9]+),([0-9]+)\]`)

	l.dev.Log("device", "try click: %s", lookup)

	retryCounter := 0
	for {
		l.dev.Log("device", "click: %s", lookup)
		xml, err := l.dev.getScreenXml()
		if err != nil {
			l.dev.Error("device", "get screen failed: %v", err)
		}
		element := xmlquery.FindOne(xml, lookup)
		if element == nil {
			if retryCounter > 3 {
				l.dev.Error("device", "Element '%s' not found", lookup)
				l.dev.pressKey(KEYCODE_BACK)
				return 0
			} else {
				time.Sleep(500 * time.Millisecond)
				retryCounter++
				continue
			}
		}

		var bounds string
		for _, attr := range element.Attr {
			if attr.Name.Local == "bounds" {
				bounds = attr.Value
				break
			}
		}

		actionContent := boundsEx.FindAllStringSubmatch(bounds, -1)
		if len(actionContent) == 0 {
			l.dev.Error("device", "No valid bounds for element '%s'", lookup)
			l.dev.pressKey(KEYCODE_BACK)
			return 0
		}
		xs, _ := strconv.ParseFloat(actionContent[0][1], 64)
		ys, _ := strconv.ParseFloat(actionContent[0][2], 64)
		xe, _ := strconv.ParseFloat(actionContent[0][3], 64)
		ye, _ := strconv.ParseFloat(actionContent[0][4], 64)

		x := (xs + xe) * 0.5
		y := (ys + ye) * 0.5

		if err := l.dev.Tap(int64(x), int64(y)); err != nil {
			l.dev.Error("device", "Touch element '%s' failed: '%v'", lookup, err)
			l.dev.pressKey(KEYCODE_BACK)
			return 0
		}
		break
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
		xml, err := l.dev.getScreenXml()
		if err != nil {
			l.dev.Error("device", "get screen failed: %v", err)
		}
		element := xmlquery.FindOne(xml, lookup)
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
	/*
		// alerts are not implemented
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
	*/

	return 0
}

func (l *LuaScriptHandler) sendKeys(L *lua.LState) int {
	content := L.CheckString(1)
	l.dev.Log("device", "send_keys")
	err := l.dev.sendText(content)
	if err != nil {
		L.RaiseError("send_keys failed: %v", err)
		return 0
	}
	return 0
}
