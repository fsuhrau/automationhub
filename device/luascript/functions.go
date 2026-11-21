package luascript

import (
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func (h *Handler) click(L *lua.LState) int {
	lookupType := L.CheckString(1)
	lookup := L.CheckString(2)
	if err := h.deps.Clicker.Click(lookupType, lookup); err != nil {
		L.RaiseError("%s", err.Error())
	}
	return 0
}

func (h *Handler) waitFor(L *lua.LState) int {
	lookupType := L.CheckString(1)
	lookup := L.CheckString(2)
	duration := L.CheckInt(3)

	h.deps.Logger.Log(logTag, "wait for: %s duration: %d", lookup, duration)
	if !h.waitForFunc(lookupType, lookup, duration) {
		L.RaiseError("Timeout: Element '%s' not found", lookup)
	}
	return 0
}

// waitForFunc polls for the element until it shows up or duration seconds passed.
func (h *Handler) waitForFunc(lookupType, lookup string, duration int) bool {
	return h.poll(duration, func() bool {
		found, err := h.deps.Finder.Find(lookupType, lookup)
		if err != nil {
			h.deps.Logger.Error(logTag, "%v", err)
		}
		return found
	})
}

func (h *Handler) exists(L *lua.LState) int {
	lookupType := L.CheckString(1)
	lookup := L.CheckString(2)
	duration := L.CheckInt(3)
	L.Push(lua.LBool(h.waitForFunc(lookupType, lookup, duration)))
	return 1
}

func (h *Handler) waitForAlertWith(L *lua.LState) int {
	contains := L.CheckString(1)
	duration := L.CheckInt(2)

	h.deps.Logger.Log(logTag, "waitForAlertWith: %s with timeout: %d", contains, duration)
	found := h.poll(duration, func() bool {
		text, err := h.deps.Alerts.AlertText()
		if err != nil {
			h.deps.Logger.Error(logTag, "get alert text failed: %v", err)
		}
		return strings.Contains(text, contains)
	})
	if !found {
		L.RaiseError("Timeout: no alert found which contains'%s'", contains)
	}
	return 0
}

func (h *Handler) sendKeys(L *lua.LState) int {
	content := L.CheckString(1)
	h.deps.Logger.Log(logTag, "send_keys")
	if err := h.deps.Keyboard.SendText(content); err != nil {
		L.RaiseError("send_keys failed: %v", err)
	}
	return 0
}

// poll runs check until it is true or the timeout in seconds passed, pausing
// pollInterval between two runs. It reports whether check succeeded.
func (h *Handler) poll(seconds int, check func() bool) bool {
	deadline := h.deps.Clock.Now().Add(time.Duration(seconds) * time.Second)
	for !check() {
		if h.deps.Clock.Now().After(deadline) {
			return false
		}
		<-h.deps.Clock.After(pollInterval)
	}
	return true
}
