package androiddevice

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/antchfx/xmlquery"
)

const screenXML = `<hierarchy>
  <node name="Kaufen" bounds="[100,200][300,400]"/>
  <node name="NoBounds"/>
  <node name="BadBounds" bounds="oops"/>
</hierarchy>`

// autoClock advances itself by the waited duration.
type autoClock struct {
	mu    sync.Mutex
	now   time.Time
	waits int
}

func (c *autoClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *autoClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	c.waits++
	ch := make(chan time.Time, 1)
	ch <- c.now
	return ch
}

type fakeScreen struct {
	xml      string
	xmlErr   error
	dumps    int
	taps     [][2]int
	tapErr   error
	keys     []int
	typed    []string
	typeErr  error
	logs     []string
	errLines []string
}

func (f *fakeScreen) getScreenXml() (*xmlquery.Node, error) {
	f.dumps++
	if f.xmlErr != nil {
		return nil, f.xmlErr
	}
	return xmlquery.Parse(strings.NewReader(f.xml))
}
func (f *fakeScreen) pressKey(key int) error { f.keys = append(f.keys, key); return nil }
func (f *fakeScreen) sendText(text string) error {
	f.typed = append(f.typed, text)
	return f.typeErr
}
func (f *fakeScreen) Tap(x, y int) error { f.taps = append(f.taps, [2]int{x, y}); return f.tapErr }
func (f *fakeScreen) Log(_, format string, args ...interface{}) {
	f.logs = append(f.logs, fmt.Sprintf(format, args...))
}
func (f *fakeScreen) Error(_, format string, args ...interface{}) {
	f.errLines = append(f.errLines, fmt.Sprintf(format, args...))
}

func run(f *fakeScreen, script string) (*autoClock, error) {
	clock := &autoClock{now: time.Unix(0, 0)}
	return clock, newScriptHandler(f, clock).Execute(script)
}

func TestNativeLuaScriptParser(t *testing.T) {
	f := &fakeScreen{xml: `<h><node name="Kaufen" bounds="[0,0][10,20]"/><node name="Anmelden" bounds="[0,0][2,2]"/><node name="OK" bounds="[0,0][4,4]"/></h>`}
	script := `
			hub.waitFor("xpath", "//*[@name='Kaufen']", 20)
			hub.click("xpath", "//*[@name='Kaufen']")
			if hub.exists("xpath", "//*[@name='Anmelden']", 10)
			then
				hub.sendKeys("xxx")
				hub.click("xpath", "//*[@name='Anmelden']")
			end
			hub.waitFor("xpath", "//*[@name='OK']", 20)
			hub.click("xpath", "//*[@name='OK']")
		`
	if _, err := run(f, script); err != nil {
		t.Fatalf("luascript error %v", err)
	}
	if len(f.taps) != 3 || f.taps[0] != [2]int{5, 10} || len(f.typed) != 1 {
		t.Errorf("taps=%v typed=%v", f.taps, f.typed)
	}
}

func TestWaitForOnTheXMLDump(t *testing.T) {
	f := &fakeScreen{xml: screenXML}
	clock, err := run(f, `hub.waitFor("xpath", "//*[@name='Missing']", 2)`)
	if err == nil || !strings.Contains(err.Error(), "Timeout: Element '//*[@name='Missing']' not found") {
		t.Fatalf("error = %v", err)
	}
	if f.dumps != 4 || clock.waits != 3 {
		t.Errorf("dumps=%d waits=%d", f.dumps, clock.waits)
	}
}

func TestFindLogsDumpFailures(t *testing.T) {
	f := &fakeScreen{xmlErr: errors.New("adb gone")}
	if _, err := run(f, `assert(not hub.exists("xpath", "//a", 0))`); err != nil {
		t.Fatal(err)
	}
	if len(f.errLines) == 0 || !strings.Contains(f.errLines[0], "get screen failed: adb gone") {
		t.Errorf("errors = %v", f.errLines)
	}
}

func TestSendKeys(t *testing.T) {
	f := &fakeScreen{typeErr: errors.New("no input")}
	_, err := run(f, `hub.sendKeys("abc")`)
	if err == nil || !strings.Contains(err.Error(), "send_keys failed: no input") {
		t.Errorf("error = %v", err)
	}
	if len(f.typed) != 1 || f.typed[0] != "abc" {
		t.Errorf("typed = %v", f.typed)
	}
}

func TestWaitForAlertWithIsNotRegistered(t *testing.T) {
	_, err := run(&fakeScreen{xml: screenXML}, `hub.waitForAlertWith("x", 1)`)
	if err == nil {
		t.Error("waitForAlertWith must not exist on Android")
	}
}

// clickWant is what a click must have done.
type clickWant struct {
	taps  [][2]int
	back  bool
	err   string
	dumps int
}

func TestClick(t *testing.T) {
	tests := []struct {
		name     string
		screen   fakeScreen
		lookup   string
		lookType string
		want     clickWant
	}{
		{name: "taps the center", lookup: "//*[@name='Kaufen']", lookType: "xpath", want: clickWant{taps: [][2]int{{200, 300}}, dumps: 1}},
		{name: "unsupported lookup type", lookup: "x", lookType: "id", want: clickWant{err: "lookupType 'id' not supported"}},
		{name: "element missing presses back after the retries", lookup: "//*[@name='Nope']", lookType: "xpath", want: clickWant{back: true, err: "Element '//*[@name='Nope']' not found", dumps: 5}},
		{name: "dump failure counts as missing", screen: fakeScreen{xmlErr: errors.New("adb")}, lookup: "//a", lookType: "xpath", want: clickWant{back: true, err: "get screen failed", dumps: 5}},
		{name: "element without bounds", lookup: "//*[@name='NoBounds']", lookType: "xpath", want: clickWant{back: true, err: "No valid bounds", dumps: 1}},
		{name: "unparsable bounds", lookup: "//*[@name='BadBounds']", lookType: "xpath", want: clickWant{back: true, err: "No valid bounds", dumps: 1}},
		{name: "tap failure presses back", screen: fakeScreen{tapErr: errors.New("input")}, lookup: "//*[@name='Kaufen']", lookType: "xpath",
			want: clickWant{taps: [][2]int{{200, 300}}, back: true, err: "Touch element '//*[@name='Kaufen']' failed: 'input'", dumps: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.screen
			f.xml = screenXML
			script := fmt.Sprintf("hub.click(%q, %q)", tt.lookType, tt.lookup)
			if _, err := run(&f, script); err != nil {
				t.Fatalf("click must not fail the script: %v", err)
			}
			assertClick(t, &f, tt.want)
		})
	}
}

func assertClick(t *testing.T, f *fakeScreen, want clickWant) {
	t.Helper()
	if fmt.Sprint(f.taps) != fmt.Sprint(want.taps) {
		t.Errorf("taps = %v, want %v", f.taps, want.taps)
	}
	if (len(f.keys) == 1 && f.keys[0] == KEYCODE_BACK) != want.back {
		t.Errorf("keys = %v, want back=%v", f.keys, want.back)
	}
	if f.dumps != want.dumps {
		t.Errorf("dumps = %d, want %d", f.dumps, want.dumps)
	}
	if want.err != "" && !strings.Contains(strings.Join(f.errLines, "\n"), want.err) {
		t.Errorf("errors = %v, want %q", f.errLines, want.err)
	}
}

func TestNewBuildsTheHandlerOfADevice(t *testing.T) {
	if New(newDevice(Deps{}.withDefaults(), nil, "S1")) == nil {
		t.Error("New returned nil")
	}
}
