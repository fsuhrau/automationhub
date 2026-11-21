package devices_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

// The device page lists what ran on a device lately: the tests of this
// project, newest first, without the runner protocols that wrap them.
func TestDeviceProtocols(t *testing.T) {
	w := newWorld(t)
	other := testkit.Project(w.DB, func(p *models.Project) { p.Identifier = "other" })
	elsewhere := testkit.Device(w.DB)
	mine := testkit.Project(w.DB, func(p *models.Project) { p.Identifier = "mine" })

	run := func(project *models.Project) uint {
		app := testkit.App(w.DB, func(a *models.App) { a.ProjectID = project.ID })
		test := testkit.Test(w.DB, func(x *models.Test) { x.AppID, x.Name = app.ID, "smoke" })
		return testkit.TestRun(w.DB, func(r *models.TestRun) { r.TestID = test.ID }).ID
	}
	protocol := func(run, dev uint, parent *uint) uint {
		return testkit.Protocol(w.DB, func(p *models.TestProtocol) {
			p.TestRunID, p.DeviceID, p.ParentTestProtocolID = run, &dev, parent
		}).ID
	}
	first, second, foreign := run(mine), run(mine), run(other)
	runner := protocol(first, w.device.ID, nil)
	a := protocol(first, w.device.ID, &runner)
	b := protocol(second, w.device.ID, nil)  // a test without a runner
	protocol(second, elsewhere.ID, nil)      // another device
	d := protocol(foreign, w.device.ID, nil) // another project's test

	get := func(project, query string) []uint {
		t.Helper()
		rec := w.Do(http.MethodGet, fmt.Sprintf("/api/%s/device/%d/protocols%s", project, w.device.ID, query), nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s%s: %d %s", project, query, rec.Code, rec.Body)
		}
		var ids []uint
		for _, p := range testkit.Decode[[]models.TestProtocol](t, rec) {
			ids = append(ids, p.ID)
		}
		return ids
	}

	if got, want := fmt.Sprint(get("mine", "")), fmt.Sprint([]uint{b, a}); got != want {
		t.Errorf("protocols %s, want %s", got, want)
	}
	if got := fmt.Sprint(get("mine", "?limit=1")); got != fmt.Sprint([]uint{b}) {
		t.Errorf("limit 1: %s", got)
	}
	if got := fmt.Sprint(get("other", "")); got != fmt.Sprint([]uint{d}) {
		t.Errorf("other project sees %s, want only its own %d", got, d)
	}
	if rec := w.Do(http.MethodGet, "/api/mine/device/999/protocols", nil); rec.Code != http.StatusNotFound {
		t.Errorf("unknown device: %d", rec.Code)
	}
}
