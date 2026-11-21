package iosdevice

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/modules/webdriver"
)

type autoClock struct{ now time.Time }

func (c *autoClock) Now() time.Time { return c.now }
func (c *autoClock) After(d time.Duration) <-chan time.Time {
	c.now = c.now.Add(d)
	ch := make(chan time.Time, 1)
	ch <- c.now
	return ch
}

type fakeLog struct{ errors []string }

func (l *fakeLog) Log(string, string, ...interface{}) {}
func (l *fakeLog) Error(_, format string, args ...interface{}) {
	l.errors = append(l.errors, fmt.Sprintf(format, args...))
}

type fakeWDA struct {
	elements   map[string]string // lookup -> element id
	emptyValue bool              // answer a found element without value
	findErr    error
	tapErr     error
	typeErr    error
	alert      string
	alertErr   error
	tapped     []string
	typed      []string
}

func (f *fakeWDA) FindElement(_, value string) (*webdriver.Element, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if f.emptyValue {
		return &webdriver.Element{}, nil
	}
	id, ok := f.elements[value]
	if !ok {
		return nil, errors.New("unable to find element")
	}
	return &webdriver.Element{Value: &webdriver.ElementValue{Element: id}}, nil
}
func (f *fakeWDA) TapElement(id string) error { f.tapped = append(f.tapped, id); return f.tapErr }
func (f *fakeWDA) SendText(text string) error { f.typed = append(f.typed, text); return f.typeErr }
func (f *fakeWDA) GetAlertText() (string, error) {
	return f.alert, f.alertErr
}

func runScript(f *fakeWDA, script string) (*fakeLog, error) {
	log := &fakeLog{}
	return log, newScriptHandler(f, log, &autoClock{now: time.Unix(0, 0)}).Execute(script)
}

func TestNativeLuaScriptParser(t *testing.T) {
	f := &fakeWDA{
		elements: map[string]string{"//*[@name='Kaufen']": "e1", "//*[@name='Anmelden']": "e2", "//*[@name='OK']": "e3"},
		alert:    "[Environment: Sandbox] confirm",
	}
	script := `
			hub.waitFor("xpath", "//*[@name='Kaufen']", 20)
			hub.click("xpath", "//*[@name='Kaufen']")
			if hub.exists("xpath", "//*[@name='Anmelden']", 10)
			then
				hub.sendKeys("xxx")
				hub.click("xpath", "//*[@name='Anmelden']")
			end
			hub.waitForAlertWith("[Environment: Sandbox]", 60)
			hub.waitFor("xpath", "//*[@name='OK']", 20)
			hub.click("xpath", "//*[@name='OK']")
		`
	if _, err := runScript(f, script); err != nil {
		t.Fatalf("luascript error %v", err)
	}
	if fmt.Sprint(f.tapped) != "[e1 e2 e3]" || fmt.Sprint(f.typed) != "[xxx]" {
		t.Errorf("tapped=%v typed=%v", f.tapped, f.typed)
	}
}

func TestScriptFailures(t *testing.T) {
	tests := []struct {
		name    string
		wda     fakeWDA
		script  string
		wantErr string
	}{
		{name: "waitFor times out", script: `hub.waitFor("xpath", "//x", 1)`, wantErr: "Timeout: Element '//x' not found"},
		{name: "click element missing", script: `hub.click("xpath", "//x")`, wantErr: "tap failed unable to find element: //x"},
		{name: "click element without value", wda: fakeWDA{emptyValue: true}, script: `hub.click("xpath", "//x")`, wantErr: "tap failed unable to find element: //x"},
		{name: "click tap fails", wda: fakeWDA{elements: map[string]string{"//x": "e"}, tapErr: errors.New("no")}, script: `hub.click("xpath", "//x")`, wantErr: "tap failed: //x"},
		{name: "sendKeys fails", wda: fakeWDA{typeErr: errors.New("no")}, script: `hub.sendKeys("a")`, wantErr: "send_keys failed: no"},
		{name: "alert never shows", wda: fakeWDA{alertErr: errors.New("no alert")}, script: `hub.waitForAlertWith("Sandbox", 1)`, wantErr: "Timeout: no alert found which contains'Sandbox'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wda := tt.wda
			_, err := runScript(&wda, tt.script)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestExistsLogsFindErrors(t *testing.T) {
	log, err := runScript(&fakeWDA{findErr: errors.New("wda down")}, `assert(not hub.exists("xpath", "//x", 0))`)
	if err != nil {
		t.Fatal(err)
	}
	if len(log.errors) == 0 || !strings.Contains(log.errors[0], "find element failed: wda down") {
		t.Errorf("errors = %v", log.errors)
	}
}

func TestFindAnswersTrueForAnElementWithoutValue(t *testing.T) {
	w := &webDriverScreen{client: &fakeWDA{emptyValue: true}}
	if found, err := w.Find("xpath", "//x"); !found || err != nil {
		t.Errorf("found=%v err=%v", found, err)
	}
}

func TestNewBuildsTheHandlerOfADevice(t *testing.T) {
	if New(nil, nil) == nil {
		t.Error("New returned nil")
	}
}
