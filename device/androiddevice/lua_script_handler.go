package androiddevice

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/fsuhrau/automationhub/device/luascript"
)

const (
	clickRetries    = 4
	clickRetryPause = 500 * time.Millisecond
)

var boundsPattern = regexp.MustCompile(`\[([0-9]+),([0-9]+)\]\[([0-9]+),([0-9]+)\]`)

// screen is what the XML dump strategy needs from the device.
type screen interface {
	getScreenXml() (*xmlquery.Node, error)
	pressKey(key int) error
	sendText(text string) error
	Tap(x, y int) error
	Log(tag, format string, args ...interface{})
	Error(tag, format string, args ...interface{})
}

// LuaScriptHandler runs hub scripts on an Android device.
type LuaScriptHandler struct {
	*luascript.Handler
}

// New creates the script handler of an Android device.
func New(dev *Device) *LuaScriptHandler {
	return newScriptHandler(dev, luascript.SystemClock())
}

func newScriptHandler(dev screen, clock luascript.Clock) *LuaScriptHandler {
	xml := &xmlScreen{dev: dev, clock: clock}
	return &LuaScriptHandler{luascript.New(luascript.Deps{
		Finder: xml, Clicker: xml, Keyboard: xml, Logger: dev, Clock: clock,
	})}
}

// xmlScreen finds and taps elements through the uiautomator XML dump.
type xmlScreen struct {
	dev   screen
	clock luascript.Clock
}

// Find reports whether the xpath lookup matches an element of the current screen.
func (s *xmlScreen) Find(_, lookup string) (bool, error) {
	node, err := s.findNode(lookup)
	return node != nil, err
}

func (s *xmlScreen) findNode(lookup string) (*xmlquery.Node, error) {
	xml, err := s.dev.getScreenXml()
	if err != nil {
		return nil, fmt.Errorf("get screen failed: %w", err)
	}
	return xmlquery.FindOne(xml, lookup), nil
}

// SendText types the text on the device.
func (s *xmlScreen) SendText(text string) error {
	return s.dev.sendText(text)
}

// Click taps the center of the element. Failures are logged and answered with
// the back key; they never fail the script.
func (s *xmlScreen) Click(lookupType, lookup string) error {
	if lookupType != "xpath" {
		s.dev.Error("device", "native script lookupType '%s' not supported", lookupType)
		return nil
	}
	s.dev.Log("device", "try click: %s", lookup)
	node := s.locate(lookup)
	if node == nil {
		s.fail("Element '%s' not found", lookup)
		return nil
	}
	x, y, ok := center(node)
	if !ok {
		s.fail("No valid bounds for element '%s'", lookup)
		return nil
	}
	if err := s.dev.Tap(x, y); err != nil {
		s.fail("Touch element '%s' failed: '%v'", lookup, err)
	}
	return nil
}

func (s *xmlScreen) locate(lookup string) *xmlquery.Node {
	for attempt := 0; attempt <= clickRetries; attempt++ {
		if attempt > 0 {
			<-s.clock.After(clickRetryPause)
		}
		s.dev.Log("device", "click: %s", lookup)
		node, err := s.findNode(lookup)
		if err != nil {
			s.dev.Error("device", "%v", err)
		}
		if node != nil {
			return node
		}
	}
	return nil
}

func (s *xmlScreen) fail(format string, args ...interface{}) {
	s.dev.Error("device", format, args...)
	_ = s.dev.pressKey(KEYCODE_BACK)
}

// center is the middle of the element's bounds attribute.
func center(node *xmlquery.Node) (x, y int, ok bool) {
	for _, attr := range node.Attr {
		if attr.Name.Local != "bounds" {
			continue
		}
		m := boundsPattern.FindStringSubmatch(attr.Value)
		if m == nil {
			return 0, 0, false
		}
		xs, ys, xe, ye := atof(m[1]), atof(m[2]), atof(m[3]), atof(m[4])
		return int((xs + xe) * 0.5), int((ys + ye) * 0.5), true
	}
	return 0, 0, false
}

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
