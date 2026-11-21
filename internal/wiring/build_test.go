package wiring_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/internal/wiring"
	"github.com/sirupsen/logrus"
)

func TestBuildMasterOpensTheDatabaseOfTheConfig(t *testing.T) {
	cfg := testConfig(t)
	cfg.Database = config.Database{SQLite: &config.SQLite{Path: t.TempDir() + "/hub.db"}}

	m, err := wiring.BuildMaster(cfg, wiring.WithLogger(quietLogger()))

	if err != nil || m.DB == nil {
		t.Fatalf("master %+v, err %v", m, err)
	}
}

func TestBuildMasterFailures(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Service
		want string
	}{
		{"no database", testConfig(t), "open database"},
		{"no folders", config.Service{}, "create folder"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := wiring.BuildMaster(tc.cfg)

			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestBuildMasterAddsTheRemoteNodeHandlerWhenEnabled(t *testing.T) {
	cfg := testConfig(t)
	cfg.DeviceManager = map[string]config.Manager{"node": {Enabled: true}}
	cfg.Hooks = []config.Hook{{Provider: "slack"}}
	var out bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&out)

	if _, err := wiring.BuildMaster(cfg, wiring.WithDB(testkit.DB(t)), wiring.WithLogger(logger)); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), "adding manager for remote nodes") {
		t.Fatalf("log %q", out.String())
	}
}

func TestBuildMasterLeavesOutDisabledNodeHandler(t *testing.T) {
	cfg := testConfig(t)
	cfg.DeviceManager = map[string]config.Manager{"node": {Enabled: false}}
	var out bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&out)

	if _, err := wiring.BuildMaster(cfg, wiring.WithDB(testkit.DB(t)), wiring.WithLogger(logger)); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(out.String(), "remote nodes") {
		t.Fatalf("log %q", out.String())
	}
}

func TestMasterAnswersPreflightsOfTheConfiguredOrigins(t *testing.T) {
	cfg := testConfig(t)
	cfg.Cors = []string{"http://ui.test"}
	m, err := wiring.BuildMaster(cfg, wiring.WithDB(testkit.DB(t)), wiring.WithLogger(quietLogger()))
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodOptions, "/api/projects", nil)
	req.Header.Set("Origin", "http://ui.test")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()
	m.Server.Handler().ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://ui.test" {
		t.Fatalf("allowed origin %q (status %d)", got, w.Code)
	}
}
