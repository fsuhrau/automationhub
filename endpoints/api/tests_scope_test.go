package api_test

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/endpoints/api/apitest"
	"github.com/fsuhrau/automationhub/endpoints/api/internal/runfixture"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/internal/testkit"
	runsvc "github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Runs and protocols are only reachable through a test of the app: a run of
// another project, or of another test, is not found.
func TestRunEndpointsRefuseOtherProjects(t *testing.T) {
	w := runfixture.New(t)
	paths := map[string]string{
		"last run of a foreign test":         w.Path("/test/%d/runs/last", w.Foreign.ID),
		"runs of a foreign test":             w.Path("/test/%d/runs", w.Foreign.ID),
		"foreign run":                        w.Path("/test/%d/run/%d", w.Foreign.ID, w.ForeignRun.ID),
		"foreign run through own test":       w.Path("/test/%d/run/%d", w.Mine.ID, w.ForeignRun.ID),
		"unknown run":                        w.Path("/test/%d/run/9999", w.Mine.ID),
		"foreign protocol":                   w.Path("/test/%d/run/%d/%d", w.Foreign.ID, w.ForeignRun.ID, w.ForeProto.ID),
		"foreign protocol through own test":  w.Path("/test/%d/run/%d/%d", w.Mine.ID, w.ForeignRun.ID, w.ForeProto.ID),
		"protocol of an unknown run":         w.Path("/test/%d/run/9999/%d", w.Mine.ID, w.Proto1.ID),
		"run of an unknown test":             w.Path("/test/9999/run/1"),
		"last run of an unknown test":        w.Path("/test/9999/runs/last"),
		"protocol of an unknown test's runs": w.Path("/test/9999/run/1/1"),
		"baselines of a foreign protocol":    w.Path("/test/%d/run/%d/%d/baselines", w.Foreign.ID, w.ForeignRun.ID, w.ForeProto.ID),
		"baseline of a foreign protocol":     w.Path("/test/%d/run/%d/%d/baseline/%d", w.Foreign.ID, w.ForeignRun.ID, w.ForeProto.ID, w.Run1.ID),
	}
	for name, path := range paths {
		t.Run(name, func(t *testing.T) {
			rec := w.Do("GET", path, nil)
			if rec.Code != http.StatusNotFound {
				t.Errorf("%s: status %d, want 404: %s", path, rec.Code, rec.Body)
			}
		})
	}
}

// The project of the path is part of the scope: the app of the other
// project is not found through this project, and a run is not cancelled
// across projects.
func TestRunEndpointsRefuseAnotherProjectsApp(t *testing.T) {
	w := runfixture.New(t)
	through := fmt.Sprintf("/api/mine/app/%d/test/%d", w.ForeignApp.ID, w.Foreign.ID)

	if rec := w.Do("GET", through+"/runs", nil); rec.Code != http.StatusNotFound {
		t.Errorf("runs of the foreign app: %d %s", rec.Code, rec.Body)
	}
	cancel := w.Path("/test/%d/run/%d/cancel", w.Foreign.ID, w.ForeignRun.ID)
	if rec := w.Do("POST", cancel, nil); rec.Code != http.StatusNotFound {
		t.Errorf("cancel of a foreign run: %d %s", rec.Code, rec.Body)
	}
}

// An id that is not a number never reaches the database.
func TestRunEndpointsRefuseNonNumericIDs(t *testing.T) {
	w := runfixture.New(t)
	bad := url.PathEscape("1 OR 1=1")
	paths := map[string]string{
		"test id of last run":     w.Path("/test/%s/runs/last", bad),
		"test id of runs":         w.Path("/test/%s/runs", bad),
		"test id of run":          w.Path("/test/%s/run/1", bad),
		"run id":                  w.Path("/test/%d/run/%s", w.Mine.ID, bad),
		"run id of protocol":      w.Path("/test/%d/run/%s/1", w.Mine.ID, bad),
		"protocol id of protocol": w.Path("/test/%d/run/%d/%s", w.Mine.ID, w.Run1.ID, bad),
		"negative run id":         w.Path("/test/%d/run/-1", w.Mine.ID),
		"protocol id of baseline": w.Path("/test/%d/run/%d/%s/baselines", w.Mine.ID, w.Run1.ID, bad),
		"baseline run id":         w.Path("/test/%d/run/%d/%d/baseline/%s", w.Mine.ID, w.Run1.ID, w.Proto1.ID, bad),
	}
	for name, path := range paths {
		t.Run(name, func(t *testing.T) {
			rec := w.Do("GET", path, nil)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s: status %d, want 400: %s", path, rec.Code, rec.Body)
			}
			if e := testkit.Decode[httpx.ErrorResponse](t, rec); e.Message == "" {
				t.Errorf("no JSON error message: %s", rec.Body)
			}
		})
	}
}

// The run of the own project answers with the same shape as before, and its
// prev/next and histories only use the own app's runs.
func TestOwnRunEndpoints(t *testing.T) {
	w := runfixture.New(t)

	last := testkit.Decode[runsvc.Page](t, w.Do("GET", w.Path("/test/%d/runs/last", w.Mine.ID), nil))
	if last.TestRun.ID != w.Run2.ID || last.PrevRunID != w.Run1.ID || last.NextRunID != 0 {
		t.Errorf("last run %d prev %d next %d", last.TestRun.ID, last.PrevRunID, last.NextRunID)
	}
	first := testkit.Decode[runsvc.Page](t, w.Do("GET", w.Path("/test/%d/run/%d", w.Mine.ID, w.Run1.ID), nil))
	if first.TestRun.ID != w.Run1.ID || first.PrevRunID != 0 || first.NextRunID != w.Run2.ID {
		t.Errorf("run %d prev %d next %d", first.TestRun.ID, first.PrevRunID, first.NextRunID)
	}
	if first.TestRun.Test == nil || len(first.TestRun.Protocols) != 1 || first.Devices == nil {
		t.Errorf("run body %+v", first)
	}
	second := testkit.Decode[runsvc.Page](t, w.Do("GET", w.Path("/test/%d/run/%d", w.Mine.ID, w.Run2.ID), nil))
	if got := second.TestRun.DeviceStatus[0].HistAvgStartupTime; got != 20 {
		t.Errorf("startup history %d, want 20 (the foreign run's 1000 must not count)", got)
	}
}

func TestOwnProtocolHistoryIgnoresOtherProjects(t *testing.T) {
	w := runfixture.New(t)

	got := testkit.Decode[models.TestRun](t, w.Do("GET", w.Path("/test/%d/run/%d/%d", w.Mine.ID, w.Run2.ID, w.Proto2.ID), nil))

	if len(got.Protocols) != 1 {
		t.Fatalf("protocols %+v", got.Protocols)
	}
	p := got.Protocols[0]
	if p.HistAvgFPS != 10 || len(p.TestProtocolHistory) != 1 || p.TestProtocolHistory[0].ID != w.Proto1.ID {
		t.Errorf("history avg %v of %d protocols, want 10 of the own earlier one", p.HistAvgFPS, len(p.TestProtocolHistory))
	}
}

// A database failure is a 500, not a 404.
func TestRunEndpointsReportDatabaseErrors(t *testing.T) {
	for _, table := range []string{"tests", "test_runs"} {
		t.Run(table, func(t *testing.T) {
			w := runfixture.New(t)
			if err := w.DB.Migrator().DropTable(table); err != nil {
				t.Fatal(err)
			}
			rec := w.Do("GET", w.Path("/test/%d/run/%d", w.Mine.ID, w.Run1.ID), nil)
			if rec.Code != http.StatusInternalServerError {
				t.Errorf("status %d, want 500: %s", rec.Code, rec.Body)
			}
		})
	}
}

// /api/data/:name serves files of the data directory only.
func TestGetDataStaysInDataPath(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "data")
	if err := os.Mkdir(data, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(data, "shot.png"), []byte("png"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "secret.txt"), []byte("secret"), 0o644)
	old := apps.TestDataPath
	apps.TestDataPath = data
	t.Cleanup(func() { apps.TestDataPath = old })
	env := apitest.New(t)

	if rec := env.Do("GET", "/api/data/shot.png", nil); rec.Code != http.StatusOK || rec.Body.String() != "png" {
		t.Errorf("own file: %d %q", rec.Code, rec.Body.String())
	}
	// a name that is not local is refused; a traversal never matches the
	// single-segment route
	for _, name := range []string{"..", "..%2fsecret.txt", "..%2f..%2fetc%2fpasswd", "a%2f..%2f..%2fsecret.txt"} {
		rec := env.Do("GET", "/api/data/"+name, nil)
		if rec.Code != http.StatusBadRequest && rec.Code != http.StatusNotFound || rec.Body.String() == "secret" {
			t.Errorf("%s: status %d %q", name, rec.Code, rec.Body.String())
		}
	}
}
