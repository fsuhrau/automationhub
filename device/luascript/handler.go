// Package luascript runs the hub Lua scripts shared by the Android and iOS devices.
// The platform specific parts (finding elements, tapping, typing, alerts) are
// supplied as small strategy interfaces.
package luascript

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

// logTag is the log tag the device handlers always used.
const logTag = "device"

// pollInterval is the pause between two lookups of a wait loop.
const pollInterval = time.Second

// Clock is the time source of the wait loops.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

// Finder looks up an element on the device screen.
type Finder interface {
	Find(lookupType, lookup string) (bool, error)
}

// Clicker taps the element a lookup describes. A returned error is raised in the script.
type Clicker interface {
	Click(lookupType, lookup string) error
}

// Keyboard types text into the focused element.
type Keyboard interface {
	SendText(text string) error
}

// Alerts reads the text of the visible system alert.
type Alerts interface {
	AlertText() (string, error)
}

// Logger writes to the device log.
type Logger interface {
	Log(tag, format string, args ...interface{})
	Error(tag, format string, args ...interface{})
}

// Deps are the strategies a Handler needs. Alerts is optional: without it the
// script function waitForAlertWith is not registered.
type Deps struct {
	Finder   Finder
	Clicker  Clicker
	Keyboard Keyboard
	Alerts   Alerts
	Logger   Logger
	Clock    Clock
}

// Handler executes hub Lua scripts against a device.
type Handler struct {
	deps Deps
}

// New creates a Handler from its strategies.
func New(deps Deps) *Handler {
	return &Handler{deps: deps}
}

// Execute runs the script. gopher-lua turns a panic inside a script function
// into the returned error, so no recover is needed here.
func (h *Handler) Execute(script string) error {
	executor := lua.NewState()
	defer executor.Close()
	h.init(executor)
	return executor.DoString(script)
}

func (h *Handler) init(L *lua.LState) int {
	functions := map[string]lua.LGFunction{
		"click":    h.click,
		"waitFor":  h.waitFor,
		"exists":   h.exists,
		"sendKeys": h.sendKeys,
	}
	if h.deps.Alerts != nil {
		functions["waitForAlertWith"] = h.waitForAlertWith
	}
	L.Push(L.RegisterModule("hub", functions))
	return 1
}

type systemClock struct{}

func (systemClock) Now() time.Time                         { return time.Now() }
func (systemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// SystemClock is the wall clock.
func SystemClock() Clock { return systemClock{} }
