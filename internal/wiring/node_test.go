package wiring_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/wiring"
)

func nodeConfig(t *testing.T) config.Service {
	t.Helper()
	cfg := testConfig(t)
	cfg.MasterURL = "127.0.0.1:1"
	return cfg
}

func TestBuildNodeWiresTheHandlersAndTheConnection(t *testing.T) {
	cfg := nodeConfig(t)
	cfg.DeviceManager = map[string]config.Manager{"android_device": {Enabled: true}}

	n, err := wiring.BuildNode(cfg, wiring.WithLogger(quietLogger()), wiring.WithDeviceHandlers(phoneHandler()))

	if err != nil || n.Server == nil || n.Handler == nil || n.Handler.ConnectFunction == nil {
		t.Fatalf("node %+v, err %v", n, err)
	}
	if dev, _ := n.Devices.GetDevice("dev-1"); dev == nil {
		t.Error("the fake handler is not in the device manager")
	}
}

func TestBuildNodeNeedsNoDatabase(t *testing.T) {
	n, err := wiring.BuildNode(nodeConfig(t))

	if err != nil || n == nil {
		t.Fatalf("node %v, err %v", n, err)
	}
}

func TestBuildNodeFailures(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Service
		want string
	}{
		{"no master", testConfig(t), "master url"},
		{"no folders", config.Service{MasterURL: "m"}, "create folder"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := wiring.BuildNode(tc.cfg)

			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
	if _, err := wiring.BuildNode(testConfig(t)); !errors.Is(err, wiring.ErrNoMasterURL) {
		t.Errorf("err = %v, want ErrNoMasterURL", err)
	}
}

func TestNodeMountsTheManagerRoutes(t *testing.T) {
	n, err := wiring.BuildNode(nodeConfig(t), wiring.WithLogger(quietLogger()))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	n.Server.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/session", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
}

func TestNodeRunReturnsTheListenError(t *testing.T) {
	cfg := nodeConfig(t)
	cfg.Port = busyPort(t)
	n, err := wiring.BuildNode(cfg, wiring.WithLogger(quietLogger()))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := n.Run(ctx); err == nil || !strings.Contains(err.Error(), "address already in use") {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildNodeFailsWhenTheBundleFolderHoldsBrokenMetadata(t *testing.T) {
	cfg := nodeConfig(t)
	if err := os.WriteFile(filepath.Join(cfg.Paths.Bundles, "broken.meta"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := wiring.BuildNode(cfg)

	if err == nil || !strings.Contains(err.Error(), "bundle store") {
		t.Fatalf("err = %v", err)
	}
}
