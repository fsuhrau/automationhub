package wiring

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device/iosdevice"
	"github.com/fsuhrau/automationhub/device/node"
	"github.com/sirupsen/logrus"
)

func TestNodeHandlersBuildsTheEnabledManagersInOrder(t *testing.T) {
	cfg := config.Service{DeviceManager: map[string]config.Manager{
		"android_device": {Enabled: true},
		"ios_sim":        {Enabled: true},
		"ios_device":     {Enabled: true},
		"macos":          {Enabled: true},
		"unity_editor":   {Enabled: true},
		"web":            {Enabled: true},
		"node":           {Enabled: true},
		"disabled":       {Enabled: false},
	}}
	var out bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&out)

	handlers := nodeHandlers(cfg, node.NewMemoryDeviceStore(cfg.DeviceManager), logger)

	var names []string
	for _, h := range handlers {
		names = append(names, h.Name())
	}
	want := "android_device,ios_sim,ios_device,macos,unity_editor,web"
	if strings.Join(names, ",") != want {
		t.Fatalf("handlers %v, want %s", names, want)
	}
	if !strings.Contains(out.String(), "adding manager unity_editor") {
		t.Errorf("log %q", out.String())
	}
}

func TestNodeHandlersSkipsDisabledAndUnknownManagers(t *testing.T) {
	cfg := config.Service{DeviceManager: map[string]config.Manager{"android_device": {Enabled: false}}}

	if got := nodeHandlers(cfg, nil, logrus.New()); len(got) != 0 {
		t.Fatalf("handlers %v", got)
	}
	if got := nodeHandlers(config.Service{}, nil, logrus.New()); len(got) != 0 {
		t.Fatalf("handlers %v", got)
	}
}

// The iOS handler reads the WebDriverAgent URL from the log output of
// go-ios, so the wiring must add its hook to the standard logger.
func TestTheIOSHandlerHookListensToTheStandardLogger(t *testing.T) {
	h := newIOSHandler(config.Manager{}, nil).(*iosdevice.Handler)

	got, err := h.Hook().Await(time.Minute, func(time.Duration) <-chan time.Time { return nil }, func(func(error)) {
		logrus.Info("ServerURLHere->http://h:9<-ServerURLHere")
	})

	if err != nil || got != "http://h:9" {
		t.Fatalf("Await = %q, %v", got, err)
	}
}
