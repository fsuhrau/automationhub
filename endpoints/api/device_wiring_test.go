package api_test

import (
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/endpoints/api/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

// The run-tests route executes the test through the unity executor the api
// builds for the device service; a pool that fails to send ends the run
// with a server error and leaves the device unlocked.
func TestDeviceRunTestsUsesTheUnityExecutor(t *testing.T) {
	env := apitest.New(t)
	testkit.Project(env.DB, func(p *models.Project) { p.Identifier = "proj" })
	testkit.Device(env.DB, func(d *models.Device) { d.DeviceIdentifier = "phone" })
	dev := testkit.NewFakeDevice("phone")
	dev.AppConnected = true
	env.Devices.Add("android_device", dev)
	env.Devices.FailWith("SendAction", http.ErrAbortHandler)

	rec := env.Do(http.MethodPost, "/api/proj/device/1/tests", map[string]string{"testName": "Suite Method", "env": "a=b"})

	if rec.Code != http.StatusInternalServerError || dev.IsLocked() {
		t.Fatalf("status = %d locked %v body %s", rec.Code, dev.IsLocked(), rec.Body)
	}
	if got := len(env.Devices.CallsTo("SendAction")); got != 2 {
		t.Fatalf("SendAction calls = %d, want the reset and the test", got)
	}
}
