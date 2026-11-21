package luascript

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// autoClock advances itself by the waited duration, so loops never sleep.
type autoClock struct {
	now   time.Time
	waits int
}

func (c *autoClock) Now() time.Time { return c.now }
func (c *autoClock) After(d time.Duration) <-chan time.Time {
	c.now = c.now.Add(d)
	c.waits++
	ch := make(chan time.Time, 1)
	ch <- c.now
	return ch
}

type fakeLog struct{ logs, errors []string }

func (l *fakeLog) Log(_, format string, args ...interface{}) {
	l.logs = append(l.logs, fmt.Sprintf(format, args...))
}
func (l *fakeLog) Error(_, format string, args ...interface{}) {
	l.errors = append(l.errors, fmt.Sprintf(format, args...))
}

// fakePlatform implements every strategy. Find answers from foundAfter calls on.
type fakePlatform struct {
	foundAfter int
	findErr    error
	finds      int
	clickErr   error
	clicked    []string
	typeErr    error
	typed      []string
	alert      string
	alertErr   error
	alertCalls int
	panicFind  bool
}

func (p *fakePlatform) Find(lookupType, lookup string) (bool, error) {
	if p.panicFind {
		panic("boom")
	}
	p.finds++
	return p.finds > p.foundAfter, p.findErr
}
func (p *fakePlatform) Click(lookupType, lookup string) error {
	p.clicked = append(p.clicked, lookupType+":"+lookup)
	return p.clickErr
}
func (p *fakePlatform) SendText(text string) error {
	p.typed = append(p.typed, text)
	return p.typeErr
}
func (p *fakePlatform) AlertText() (string, error) {
	p.alertCalls++
	return p.alert, p.alertErr
}

func setup(p *fakePlatform, withAlerts bool) (*Handler, *fakeLog, *autoClock) {
	log := &fakeLog{}
	clock := &autoClock{now: time.Unix(0, 0)}
	deps := Deps{Finder: p, Clicker: p, Keyboard: p, Logger: log, Clock: clock}
	if withAlerts {
		deps.Alerts = p
	}
	return New(deps), log, clock
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name       string
		platform   fakePlatform
		script     string
		wantErr    string
		wantFinds  int
		wantWaits  int
		wantClicks []string
		wantTyped  []string
	}{
		{name: "waitFor finds at once", script: `hub.waitFor("xpath", "//a", 5)`, wantFinds: 1},
		{name: "waitFor finds after retries", platform: fakePlatform{foundAfter: 2}, script: `hub.waitFor("xpath", "//a", 5)`, wantFinds: 3, wantWaits: 2},
		{name: "waitFor times out", platform: fakePlatform{foundAfter: 100}, script: `hub.waitFor("xpath", "//a", 2)`,
			wantErr: "Timeout: Element '//a' not found", wantFinds: 4, wantWaits: 3},
		{name: "waitFor logs finder errors and keeps polling", platform: fakePlatform{foundAfter: 1, findErr: errors.New("x")},
			script: `hub.waitFor("xpath", "//a", 5)`, wantFinds: 2, wantWaits: 1},
		{name: "exists true", script: `assert(hub.exists("xpath", "//a", 1))`, wantFinds: 1},
		{name: "exists false", platform: fakePlatform{foundAfter: 100}, script: `assert(not hub.exists("xpath", "//a", 1))`, wantFinds: 3, wantWaits: 2},
		{name: "click", script: `hub.click("xpath", "//a")`, wantClicks: []string{"xpath://a"}},
		{name: "click error is raised", platform: fakePlatform{clickErr: errors.New("tap failed: //a")},
			script: `hub.click("xpath", "//a")`, wantErr: "tap failed: //a", wantClicks: []string{"xpath://a"}},
		{name: "sendKeys", script: `hub.sendKeys("abc")`, wantTyped: []string{"abc"}},
		{name: "sendKeys error is raised", platform: fakePlatform{typeErr: errors.New("nope")},
			script: `hub.sendKeys("abc")`, wantErr: "send_keys failed: nope", wantTyped: []string{"abc"}},
		{name: "missing argument", script: `hub.click("xpath")`, wantErr: "bad argument #2"},
		{name: "syntax error", script: `hub.`, wantErr: "syntax error"},
		{name: "panic becomes an error", platform: fakePlatform{panicFind: true}, script: `hub.waitFor("xpath", "//a", 1)`, wantErr: "boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.platform
			h, _, clock := setup(&p, false)
			err := h.Execute(tt.script)
			assertError(t, err, tt.wantErr)
			if p.finds != tt.wantFinds || clock.waits != tt.wantWaits {
				t.Errorf("finds=%d waits=%d, want %d/%d", p.finds, clock.waits, tt.wantFinds, tt.wantWaits)
			}
			if fmt.Sprint(p.clicked) != fmt.Sprint(tt.wantClicks) || fmt.Sprint(p.typed) != fmt.Sprint(tt.wantTyped) {
				t.Errorf("clicked=%v typed=%v", p.clicked, p.typed)
			}
		})
	}
}

func TestExecuteLogsFinderErrors(t *testing.T) {
	p := &fakePlatform{foundAfter: 1, findErr: errors.New("screen gone")}
	h, log, _ := setup(p, false)
	_ = h.Execute(`hub.waitFor("xpath", "//a", 5)`)
	if len(log.errors) != 2 || log.errors[0] != "screen gone" {
		t.Errorf("errors = %v", log.errors)
	}
}

func TestWaitForAlertWith(t *testing.T) {
	tests := []struct {
		name      string
		platform  fakePlatform
		wantErr   string
		wantCalls int
	}{
		{name: "alert matches", platform: fakePlatform{alert: "Sandbox alert"}, wantCalls: 1},
		{name: "alert never matches", platform: fakePlatform{alert: "other"},
			wantErr: "Timeout: no alert found which contains'Sandbox'", wantCalls: 3},
		{name: "alert error is logged and times out", platform: fakePlatform{alertErr: errors.New("none")},
			wantErr: "Timeout: no alert found", wantCalls: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.platform
			h, log, _ := setup(&p, true)
			err := h.Execute(`hub.waitForAlertWith("Sandbox", 1)`)
			assertError(t, err, tt.wantErr)
			if p.alertCalls != tt.wantCalls {
				t.Errorf("alert calls = %d, want %d", p.alertCalls, tt.wantCalls)
			}
			if tt.platform.alertErr != nil && len(log.errors) == 0 {
				t.Error("alert error was not logged")
			}
		})
	}
}

func TestWaitForAlertWithIsOnlyRegisteredWithAlerts(t *testing.T) {
	h, _, _ := setup(&fakePlatform{}, false)
	err := h.Execute(`hub.waitForAlertWith("x", 1)`)
	if err == nil || !strings.Contains(err.Error(), "attempt to call") {
		t.Errorf("error = %v, want a call of nil", err)
	}
}

func TestRunningTheHubModule(t *testing.T) {
	p := &fakePlatform{alert: "[Environment: Sandbox]", foundAfter: 1}
	h, _, _ := setup(p, true)
	script := `
		hub.waitFor("xpath", "//*[@name='Kaufen']", 20)
		hub.click("xpath", "//*[@name='Kaufen']")
		if hub.exists("xpath", "//*[@name='Anmelden']", 10) then
			hub.sendKeys("xxx")
			hub.click("xpath", "//*[@name='Anmelden']")
		end
		hub.waitForAlertWith("[Environment: Sandbox]", 60)`
	if err := h.Execute(script); err != nil {
		t.Fatal(err)
	}
	if len(p.clicked) != 2 || len(p.typed) != 1 {
		t.Errorf("clicked=%v typed=%v", p.clicked, p.typed)
	}
}

func TestSystemClock(t *testing.T) {
	c := SystemClock()
	if time.Since(c.Now()) > time.Minute {
		t.Error("Now is not the current time")
	}
	select {
	case <-c.After(time.Millisecond):
	case <-time.After(5 * time.Second):
		t.Error("After did not fire")
	}
}

func assertError(t *testing.T, err error, want string) {
	t.Helper()
	if want == "" && err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want != "" && (err == nil || !strings.Contains(err.Error(), want)) {
		t.Fatalf("error = %v, want containing %q", err, want)
	}
}
