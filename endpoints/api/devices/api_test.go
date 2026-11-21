package devices_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/internal/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

const base = "/api/proj"

// world is an api with the project "proj" and the device "phone".
type world struct {
	*apitest.Env
	device *models.Device
}

func newWorld(t *testing.T, opts ...testkit.Option[models.Device]) *world {
	t.Helper()
	env := apitest.New(t)
	testkit.Project(env.DB, func(p *models.Project) { p.Identifier = "proj" })
	opts = append([]testkit.Option[models.Device]{func(d *models.Device) {
		d.DeviceIdentifier, d.Alias = "phone", "Old alias"
		d.CustomParameter = []models.CustomParameter{{Key: "PIN", Value: "1234"}, {Key: "LANG", Value: "de"}, {Key: "GONE", Value: "x"}}
	}}, opts...)
	return &world{Env: env, device: testkit.Device(env.DB, opts...)}
}

func (w *world) path(suffix string) string {
	return fmt.Sprintf("%s/device/%d%s", base, w.device.ID, suffix)
}

func dropDevices(t *testing.T, w *world) {
	t.Helper()
	if err := w.DB.Migrator().DropTable(&models.Device{}); err != nil {
		t.Fatal(err)
	}
}

// A failing query is a server error, not "not found": the list answered 404
// when the database failed.
func TestListDatabaseFailureIs500(t *testing.T) {
	w := newWorld(t)
	dropDevices(t, w)

	rec := w.Do(http.MethodGet, base+"/devices", nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
}

// A request that is not a test signature is a bad request, not "not found":
// the run-tests route answered 404 for it.
func TestRunTestsBadInputIs400(t *testing.T) {
	tests := []struct{ name, body string }{
		{"one word", `{"testName":"x"}`},
		{"three words", `{"testName":"a b c"}`},
		{"no test name", `{}`},
		{"not json", `testName=x`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newWorld(t)
			req := httptest.NewRequest(http.MethodPost, w.path("/tests"), strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")

			rec := w.Request(req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
			}
		})
	}
}

func TestRunTestsStatusCodesOfTheOtherFailures(t *testing.T) {
	body := map[string]string{"testName": "Suite Method"}
	w := newWorld(t)
	if rec := w.Do(http.MethodPost, base+"/device/99/tests", body); rec.Code != http.StatusNotFound {
		t.Errorf("unknown device: %d", rec.Code)
	}
	if rec := w.Do(http.MethodPost, w.path("/tests"), body); rec.Code != http.StatusNotFound {
		t.Errorf("device not connected: %d %s", rec.Code, rec.Body)
	}
	dev := testkit.NewFakeDevice("phone")
	w.Devices.Add("android_device", dev)
	if rec := w.Do(http.MethodPost, w.path("/tests"), body); rec.Code != http.StatusNotFound || dev.IsLocked() {
		t.Errorf("no app connected: %d %s", rec.Code, rec.Body)
	}
	dev.AppConnected = true
	w.Devices.FailWith("SendAction", fmt.Errorf("device gone"))
	if rec := w.Do(http.MethodPost, w.path("/tests"), body); rec.Code != http.StatusInternalServerError || dev.IsLocked() {
		t.Errorf("execution fails: %d %s", rec.Code, rec.Body)
	}
}

func TestListFiltersByPlatform(t *testing.T) {
	w := newWorld(t, func(d *models.Device) { d.PlatformType = models.PlatformTypeAndroid })
	testkit.Device(w.DB, func(d *models.Device) { d.PlatformType = models.PlatformTypeiOS })

	all := testkit.Decode[[]models.Device](t, w.Do(http.MethodGet, base+"/devices", nil))
	android := testkit.Decode[[]models.Device](t, w.Do(http.MethodGet, fmt.Sprintf("%s/devices?platform=%d", base, models.PlatformTypeAndroid), nil))

	if len(all) != 2 || len(android) != 1 || android[0].ID != w.device.ID {
		t.Fatalf("all %d android %+v", len(all), android)
	}
}

func TestGetReportsTheLiveState(t *testing.T) {
	w := newWorld(t)
	live := testkit.NewFakeDevice("phone")
	live.State = 2
	w.Devices.Add("android_device", live)

	got := testkit.Decode[models.Device](t, w.Do(http.MethodGet, w.path(""), nil))

	if got.Status != live.State || len(got.CustomParameter) != 3 {
		t.Fatalf("got %+v", got)
	}
}

func TestUnknownAndForeignDevicesAreNotFound(t *testing.T) {
	w := newWorld(t, func(d *models.Device) { d.CompanyID = 2 })
	testkit.Project(w.DB, func(p *models.Project) { p.Identifier = "mine"; p.CompanyID = 1 })
	foreign := w.path("")
	routes := []struct{ method, path string }{
		{"GET", base + "/device/99"}, {"GET", foreign}, {"DELETE", foreign}, {"PUT", foreign},
		{"POST", foreign + "/unlock"}, {"GET", foreign + "/protocols"},
	}
	for _, r := range routes {
		path := strings.Replace(r.path, "/proj/", "/mine/", 1)
		if rec := w.Do(r.method, path, map[string]string{}); rec.Code != http.StatusNotFound {
			t.Errorf("%s %s: %d", r.method, path, rec.Code)
		}
	}
}

func TestDeleteRemovesTheDevice(t *testing.T) {
	w := newWorld(t)

	if rec := w.Do(http.MethodDelete, w.path(""), nil); rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec := w.Do(http.MethodGet, w.path(""), nil); rec.Code != http.StatusNotFound {
		t.Fatalf("after delete: %d", rec.Code)
	}
}

func TestUnlockReleasesALockedDevice(t *testing.T) {
	w := newWorld(t)
	if rec := w.Do(http.MethodPost, w.path("/unlock"), nil); rec.Code != http.StatusNotFound {
		t.Fatalf("not connected: %d", rec.Code)
	}
	dev := testkit.NewFakeDevice("phone")
	_ = dev.Lock()
	w.Devices.Add("android_device", dev)

	rec := w.Do(http.MethodPost, w.path("/unlock"), nil)

	if rec.Code != http.StatusOK || dev.IsLocked() {
		t.Fatalf("status = %d locked %v", rec.Code, dev.IsLocked())
	}
	dev.FailWith("Unlock", fmt.Errorf("stuck"))
	_ = dev.Lock()
	if rec := w.Do(http.MethodPost, w.path("/unlock"), nil); rec.Code != http.StatusConflict {
		t.Fatalf("unlock fails: %d", rec.Code)
	}
}

func (w *world) paramsOf(deviceID uint) string {
	var list []models.CustomParameter
	w.DB.Where("device_id = ?", deviceID).Find(&list)
	var out []string
	for _, p := range list {
		out = append(out, p.Key+"="+p.Value)
	}
	sort.Strings(out)
	return strings.Join(out, " ")
}

// The edit page changes the alias and the custom parameters. Before, a
// renamed key deleted its parameter, edited values weren't saved (GORM only
// upserts the foreign key of existing associations), and an empty alias was
// ignored.
func TestUpdateDevice(t *testing.T) {
	w := newWorld(t)
	other := testkit.Device(w.DB, func(d *models.Device) { d.CustomParameter = []models.CustomParameter{{Key: "OTHER", Value: "o"}} })
	pin, lang := w.device.CustomParameter[0].ID, w.device.CustomParameter[1].ID

	body := fmt.Sprintf(`{"alias":"  ","customParameter":[
		{"id":%d,"key":"PIN","value":"9999"},
		{"id":%d,"key":" LOCALE ","value":"en"},
		{"key":"NEW","value":"n"},
		{"id":%d,"key":"STOLEN","value":"s"}]}`, pin, lang, other.CustomParameter[0].ID)
	rec := w.Request(jsonRequest(http.MethodPut, w.path(""), body))

	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if got, want := w.paramsOf(w.device.ID), "LOCALE=en NEW=n PIN=9999 STOLEN=s"; got != want {
		t.Errorf("parameters %q, want %q", got, want)
	}
	if got := w.paramsOf(other.ID); got != "OTHER=o" {
		t.Errorf("the other device's parameters changed: %q", got)
	}
	var saved models.Device
	w.DB.First(&saved, w.device.ID)
	if saved.Alias != "" {
		t.Errorf("alias %q, want it cleared", saved.Alias)
	}
	if resp := testkit.Decode[models.Device](t, rec); len(resp.CustomParameter) != 4 {
		t.Errorf("response %s", rec.Body)
	}
	var renamed models.CustomParameter
	w.DB.First(&renamed, lang)
	if renamed.Key != "LOCALE" {
		t.Errorf("renamed parameter %+v, want the same row with the new key", renamed)
	}
}

func TestUpdateRejectsBadBodiesWithoutChangingAnything(t *testing.T) {
	w := newWorld(t)
	before := w.paramsOf(w.device.ID)
	for name, bad := range map[string]string{
		"empty key":     `{"customParameter":[{"key":" ","value":"v"}]}`,
		"duplicate key": `{"customParameter":[{"key":"A","value":"1"},{"key":"A","value":"2"}]}`,
		"not JSON":      `alias=x`,
	} {
		if rec := w.Request(jsonRequest(http.MethodPut, w.path(""), bad)); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", name, rec.Code, rec.Body)
		}
	}
	if got := w.paramsOf(w.device.ID); got != before {
		t.Errorf("a rejected update changed the parameters: %q", got)
	}
}

func jsonRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestControlRoutesResolveTheDevice(t *testing.T) {
	w := newWorld(t)
	if rec := w.Do(http.MethodGet, w.path("/lock"), nil); rec.Code != http.StatusConflict {
		t.Fatalf("not connected: %d %s", rec.Code, rec.Body)
	}
	w.Devices.Add("android_device", testkit.NewFakeDevice("phone"))
	if rec := w.Do(http.MethodGet, w.path("/lock"), nil); rec.Code != http.StatusOK {
		t.Fatalf("connected: %d %s", rec.Code, rec.Body)
	}
	if rec := w.Do(http.MethodGet, base+"/device/99/lock", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown: %d", rec.Code)
	}
}
