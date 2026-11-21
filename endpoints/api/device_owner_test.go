package api_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/internal/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Two different users with the same display name are two owners: the second
// is refused and learns nothing of the lease of the first.
func TestUsersWithTheSameNameDoNotShareALease(t *testing.T) {
	env := apitest.New(t)
	testkit.Project(env.DB, func(p *models.Project) { p.Identifier = "proj" })
	testkit.Device(env.DB, func(d *models.Device) { d.DeviceIdentifier = "phone" })
	env.Devices.Add("android_device", testkit.NewFakeDevice("phone"))
	const lock = "/api/proj/device/1/lock"

	env.SignInAs(authentication.User{ID: 1, Name: "admin"})
	first := env.Do(http.MethodPost, lock, nil)
	if first.Code != http.StatusOK {
		t.Fatalf("first lock: %d %s", first.Code, first.Body)
	}
	token := testkit.Decode[control.Lease](t, first).Token

	env.SignInAs(authentication.User{ID: 2, Name: "admin"})
	second := env.Do(http.MethodPost, lock, nil)
	if second.Code != http.StatusConflict || strings.Contains(second.Body.String(), token) {
		t.Fatalf("second user: %d %s", second.Code, second.Body)
	}
	env.SignInAs(authentication.User{ID: 1, Name: "admin"})
	if again := env.Do(http.MethodPost, lock, nil); again.Code != http.StatusOK {
		t.Fatalf("the owner lost the lease: %d %s", again.Code, again.Body)
	}
}
