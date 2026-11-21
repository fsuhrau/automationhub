package control

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
)

func TestAppLifecycle(t *testing.T) {
	e := setup(t)
	token := e.lock(t, "1", "alice")
	body := map[string]any{"binaryId": 7, "clearData": true}

	if rec := e.call(t, token, http.MethodPost, "1/app/install", body); rec.Code != http.StatusNoContent {
		t.Fatalf("install: %d %s", rec.Code, rec.Body)
	}
	rec := e.call(t, token, http.MethodPost, "1/app/start", body)
	if rec.Code != http.StatusOK || testkit.Decode[map[string]string](t, rec)["sessionId"] == "" {
		t.Fatalf("start: %d %s", rec.Code, rec.Body)
	}
	if rec := e.call(t, token, http.MethodPost, "1/app/stop", body); rec.Code != http.StatusNoContent {
		t.Fatalf("stop: %d %s", rec.Code, rec.Body)
	}
	for _, method := range []string{"InstallApp", "StartApp", "StopApp"} {
		if len(e.full.CallsTo(method)) != 1 {
			t.Fatalf("%s not called once: %+v", method, e.full.Calls())
		}
	}
	if len(e.full.CallsTo("ClearAppData")) != 1 {
		t.Fatal("data not cleared on stop")
	}
}

func TestAppCallsNeedValidInput(t *testing.T) {
	e := setup(t)
	token := e.lock(t, "1", "alice")
	if rec := e.call(t, token, http.MethodPost, "1/app/install", map[string]any{}); rec.Code != http.StatusBadRequest {
		t.Fatalf("without binary: expected 400, got %d", rec.Code)
	}
	e.resolveApp = func(*gin.Context, uint) (*app.Parameter, error) { return nil, errors.New("no such bundle") }
	rec := e.call(t, token, http.MethodPost, "1/app/install", map[string]any{"binaryId": 9})
	if rec.Code != http.StatusNotFound || testkit.Decode[map[string]string](t, rec)["message"] != "no such bundle" {
		t.Fatalf("unknown bundle: %d %s", rec.Code, rec.Body)
	}
}

func TestAppCallsMapServiceErrors(t *testing.T) {
	tests := []struct {
		name, path string
		token      string
		fail       string
		want       int
	}{
		{"install without lease", "1/app/install", "", "", http.StatusPreconditionRequired},
		{"start without lease", "1/app/start", "", "", http.StatusPreconditionRequired},
		{"stop without lease", "1/app/stop", "", "", http.StatusPreconditionRequired},
		{"install fails", "1/app/install", "lease", "InstallApp", http.StatusInternalServerError},
		{"start fails", "1/app/start", "lease", "StartApp", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setup(t)
			token := ""
			if tt.token != "" {
				token = e.lock(t, "1", "alice")
			}
			e.full.FailWith(tt.fail, errors.New("boom"))
			if rec := e.call(t, token, http.MethodPost, tt.path, map[string]any{"binaryId": 1}); rec.Code != tt.want {
				t.Fatalf("expected %d, got %d %s", tt.want, rec.Code, rec.Body)
			}
		})
	}
}

func TestAppCallsRejectABadBody(t *testing.T) {
	for _, path := range []string{"1/app/install", "1/app/start", "1/app/stop"} {
		t.Run(path, func(t *testing.T) {
			e := setup(t)
			if rec := e.call(t, "", http.MethodPost, path, json.RawMessage(`[]`)); rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d %s", rec.Code, rec.Body)
			}
		})
	}
}
