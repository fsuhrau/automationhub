package tests_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/internal/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	service "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

// world is the project proj with app 1, one Unity test and two devices; a
// second project owns a foreign test.
type world struct {
	env           *apitest.Env
	test, foreign *models.Test
	devices       []*models.Device
	path          string
}

func newWorld(t *testing.T) *world {
	t.Helper()
	env := apitest.New(t)
	db := env.DB
	project := testkit.Project(db, func(p *models.Project) { p.Identifier, p.CompanyID = "proj", 1 })
	app := testkit.App(db, func(a *models.App) { a.ProjectID = project.ID })
	w := &world{env: env, path: fmt.Sprintf("/api/proj/app/%d", app.ID)}
	for _, company := range []uint{0, 99} {
		w.devices = append(w.devices, testkit.Device(db, func(d *models.Device) { d.CompanyID = company }))
	}
	w.devices = append(w.devices, testkit.Device(db))
	w.test = testkit.Test(db, func(x *models.Test) { x.AppID, x.Name = app.ID, "Smoke" })
	db.Create(&models.TestConfig{TestID: w.test.ID, Type: models.TestTypeUnity, AllDevices: true})
	db.Create(&models.TestConfigUnity{TestConfigID: 1})
	other := testkit.App(db)
	w.foreign = testkit.Test(db, func(x *models.Test) { x.AppID = other.ID })
	return w
}

func (w *world) testPath() string { return fmt.Sprintf("%s/test/%d", w.path, w.test.ID) }

func decode[T any](t *testing.T, w *world, method, path string, body any, wantStatus int) T {
	t.Helper()
	rec := w.env.Do(method, path, body)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s: status %d, want %d: %s", method, path, rec.Code, wantStatus, rec.Body)
	}
	return testkit.Decode[T](t, rec)
}

func TestListShowsTheAppsTestsWithRecentRuns(t *testing.T) {
	w := newWorld(t)
	testkit.TestRun(w.env.DB, func(r *models.TestRun) { r.TestID = w.test.ID })

	list := decode[[]service.ListEntry](t, w, "GET", w.path+"/tests", nil, http.StatusOK)

	if len(list) != 1 || list[0].Name != "Smoke" || len(list[0].RecentRuns) != 1 || list[0].RecentRuns[0].Status != "empty" {
		t.Fatalf("list %+v", list)
	}
}

func TestGetAnswersTheTestOrNotFound(t *testing.T) {
	w := newWorld(t)
	tests := []struct {
		name   string
		path   string
		status int
	}{
		{"own test", w.testPath(), http.StatusOK},
		{"unknown test", w.path + "/test/99", http.StatusNotFound},
		{"another app's test", fmt.Sprintf("%s/test/%d", w.path, w.foreign.ID), http.StatusNotFound},
		{"non-numeric id", w.path + "/test/abc", http.StatusNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := w.env.Do("GET", tc.path, nil)

			if rec.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
			if tc.status == http.StatusNotFound && testkit.Decode[httpx.ErrorResponse](t, rec).Message != "record not found" {
				t.Fatalf("body %s", rec.Body)
			}
		})
	}
}

func TestCreateStoresANewUnityTest(t *testing.T) {
	w := newWorld(t)
	body := map[string]any{"type": 0, "name": " Nightly ", "allDevices": false, "devices": []uint{w.devices[0].ID}, "unityTestCategoryType": 0}

	created := decode[models.Test](t, w, "POST", w.path+"/test", body, http.StatusCreated)

	got := decode[models.Test](t, w, "GET", fmt.Sprintf("%s/test/%d", w.path, created.ID), nil, http.StatusOK)
	if got.Name != "Nightly" || len(got.TestConfig.Devices) != 1 {
		t.Fatalf("stored %+v", got)
	}
}

func TestCreateRefusals(t *testing.T) {
	w := newWorld(t)
	tests := []struct {
		name string
		body any
		want string
	}{
		{"not json", "name=x", "json: cannot unmarshal"},
		{"other test type", map[string]any{"type": 1, "name": "x"}, "only Unity tests can be created here"},
		{"no name", map[string]any{"type": 0, "name": " ", "allDevices": true}, "the test needs a name"},
		{"foreign company's device", map[string]any{"type": 0, "name": "x", "devices": []uint{w.devices[1].ID}}, "some of the picked devices don't exist"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := w.env.Do("POST", w.path+"/test", tc.body)

			msg := testkit.Decode[httpx.ErrorResponse](t, rec).Message
			if rec.Code != http.StatusBadRequest || len(msg) < len(tc.want) || msg[:len(tc.want)] != tc.want {
				t.Fatalf("%d %q, want 400 starting %q", rec.Code, msg, tc.want)
			}
		})
	}
}

func TestUpdateSavesTheEditPage(t *testing.T) {
	w := newWorld(t)
	body := map[string]any{"name": " Smoke 2 ", "devices": []uint{w.devices[0].ID, w.devices[2].ID}, "recordSession": true,
		"unityTestCategoryType": 1, "categories": "a, a"}

	saved := decode[models.Test](t, w, "PUT", w.testPath(), body, http.StatusOK)

	if saved.Name != "Smoke 2" || len(saved.TestConfig.Devices) != 2 || !saved.TestConfig.RecordSession || saved.TestConfig.Unity.Categories != "a" {
		t.Fatalf("saved %+v", saved)
	}
}

func TestUpdateRefusals(t *testing.T) {
	w := newWorld(t)
	foreign := fmt.Sprintf("%s/test/%d", w.path, w.foreign.ID)
	tests := []struct {
		name   string
		path   string
		body   any
		status int
	}{
		{"not json", w.testPath(), "name=x", http.StatusBadRequest},
		{"no name", w.testPath(), map[string]any{"name": " "}, http.StatusBadRequest},
		{"another app's test", foreign, map[string]any{"name": "x", "allDevices": true}, http.StatusNotFound},
		{"non-numeric id", w.path + "/test/abc", map[string]any{"name": "x", "allDevices": true}, http.StatusNotFound},
		{"foreign company's device", w.testPath(), map[string]any{"name": "x", "devices": []uint{w.devices[1].ID}}, http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := w.env.Do("PUT", tc.path, tc.body)

			if rec.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
		})
	}
}

func TestDatabaseFailuresAreServerErrors(t *testing.T) {
	w := newWorld(t)
	if err := w.env.DB.Migrator().DropTable("test_configs"); err != nil {
		t.Fatal(err)
	}

	for _, req := range [][2]string{{"GET", "/tests"}, {"GET", fmt.Sprintf("/test/%d", w.test.ID)}} {
		rec := w.env.Do(req[0], w.path+req[1], nil)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("%v: status %d, want 500: %s", req, rec.Code, rec.Body)
		}
	}
}

// A duplicate key in the storage is a conflict (409), not a server error
// (500). The hub's schema has no such key for a test's name; the test adds
// one to make the database refuse the second row.
func TestCreateAnswersAConflictWhenTheDatabaseRefusesADuplicate(t *testing.T) {
	w := newWorld(t)
	w.env.DB.Exec("CREATE UNIQUE INDEX one_name ON tests (app_id, name)")
	body := map[string]any{"type": 0, "name": "Smoke", "allDevices": true}

	rec := w.env.Do("POST", w.path+"/test", body)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status %d, want 409: %s", rec.Code, rec.Body)
	}
}
