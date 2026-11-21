package web

import (
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device/hostinfo"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

const (
	uuid      = "54DECB62-3993-4031-9C6A-18CE048CC63C"
	chromeBin = "/Applications/Chrome"
	safariApp = "/Applications/Safari.app"
	pageURL   = "http://game.local/play"
)

type testEnv struct {
	env    hostinfo.Env
	runner *testkit.FakeRunner
}

func newTestEnv(goos string) *testEnv {
	r := &testkit.FakeRunner{}
	clock := testkit.NewFakeClock(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	return &testEnv{env: hostinfo.Env{Runner: r, Clock: clock, GOOS: goos}, runner: r}
}

// scriptMac scripts the host commands and the browser version commands of a Mac.
func (e *testEnv) scriptMac() {
	out := func(line, text string) { e.runner.Script(line, []byte(text), nil) }
	out("sw_vers -productName", "macOS\n")
	out("sw_vers -productVersion", "14.5\n")
	out("sysctl -n kern.hostname", "mac-1\n")
	out("ioreg -rd1 -c IOPlatformExpertDevice", "    \"IOPlatformUUID\" = \""+uuid+"\"\n")
	out("system_profiler SPHardwareDataType", "Model Identifier: Mac14,2\nSerial Number (system): C02XYZ\n")
	out(chromeBin+" --version", "Google Chrome 126.0.1\n")
	out("defaults read "+safariApp+"/Contents/Info CFBundleShortVersionString", "17.5\n")
}

func webParams() *app.Parameter { return &app.Parameter{Web: &app.WebParams{StartURL: pageURL}} }
