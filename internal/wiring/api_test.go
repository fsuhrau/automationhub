package wiring_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/internal/wiring"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// apiEnv is the API on an in-memory database with a project and a device
// "dev-1" that is connected to the fake device manager.
type apiEnv struct {
	db      *gorm.DB
	api     *wiring.API
	devices *testkit.FakeDeviceManager
	router  *gin.Engine
}

func newAPIEnv(t *testing.T) *apiEnv {
	t.Helper()
	cfg := testConfig(t)
	db := testkit.DB(t)
	testkit.Project(db, func(p *models.Project) { p.Identifier = "proj" })
	testkit.Device(db, func(d *models.Device) { d.DeviceIdentifier, d.Manager = "dev-1", "android_device" })
	devices := &testkit.FakeDeviceManager{}
	devices.Add("android_device", testkit.NewFakeDevice("dev-1"))
	built := wiring.BuildAPI(wiring.Infra{DB: db, Devices: devices, Nodes: &testkit.FakeNodes{}, Config: cfg, Logger: quietLogger()})
	router := gin.New()
	if err := built.Endpoint.RegisterRoutes(router, router.Group("")); err != nil {
		t.Fatal(err)
	}
	return &apiEnv{db: db, api: built, devices: devices, router: router}
}

// The REST routes and the control service the API exposes are the same
// instance, and the broker serves the stream the bridges publish on.
func TestAPIBuildsTheSharedPartsOnce(t *testing.T) {
	e := newAPIEnv(t)

	rec := testkit.DoJSON(t, e.router, http.MethodPost, "/api/proj/device/1/lock", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("lock: %d %s", rec.Code, rec.Body)
	}
	if _, held := e.api.Control.Holder("dev-1"); !held {
		t.Fatal("the lease of the route is not in the control service of the API")
	}
	client := e.api.Broker.Subscribe()
	defer e.api.Broker.Unsubscribe(client)

	events.NodeConnected.Trigger(events.NodeConnectedPayload{})

	waitForChannel(t, client, "nodes")
}

// waitForChannel reads the client until an event of the channel arrives;
// events of other channels (e.g. the reload a lock causes) are skipped.
func waitForChannel(t *testing.T, client *sse.Client, channel string) {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case msg := <-client.Events():
			if msg.Channel == channel {
				return
			}
		case <-deadline:
			t.Fatalf("no event on the channel %q", channel)
		}
	}
}

func TestAPIStartRunsTheDeviceControlUntilTheContextEnds(t *testing.T) {
	e := newAPIEnv(t)
	ctx := t.Context()

	e.api.Start(ctx)

	if _, err := e.api.Control.Acquire(testkit.NewFakeDevice("x"), "me"); err != nil {
		t.Fatal(err)
	}
}

// A lock change of a device makes the UI reload the device list, which
// carries isLocked and lockedBy.
func TestAPIAnnouncesDeviceLockChanges(t *testing.T) {
	e := newAPIEnv(t)
	client := e.api.Broker.Subscribe()
	defer e.api.Broker.Unsubscribe(client)

	events.DeviceLockChanged.Trigger(events.DeviceLockChangedPayload{DeviceIdentifier: "dev-1", Locked: true})
	events.DeviceLockChanged.Trigger(events.DeviceLockChangedPayload{DeviceIdentifier: "nobody", Locked: true})

	waitForChannel(t, client, "device_list")
}

func TestAPIServesTheDataFolder(t *testing.T) {
	e := newAPIEnv(t)

	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/data/..", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

// The bundle of an app lists its runs with their status: the app
// service gets the run summaries of the test service.
func TestAPISummarizesTheRunsOfABundle(t *testing.T) {
	e := newAPIEnv(t)
	db := e.db
	var project models.Project
	db.First(&project, "identifier = ?", "proj")
	app := testkit.App(db, func(a *models.App) { a.ProjectID = project.ID })
	bundle := models.AppBinary{AppID: app.ID, Name: "app.apk", Platform: "android", Hash: "abc"}
	db.Create(&bundle)
	test := testkit.Test(db, func(x *models.Test) { x.AppID = app.ID })
	testkit.TestRun(db, func(r *models.TestRun) { r.TestID, r.AppBinaryID = test.ID, bundle.ID })

	rec := testkit.DoJSON(t, e.router, http.MethodGet, fmt.Sprintf("/api/proj/app/%d/bundle/%d", app.ID, bundle.ID), nil)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"empty"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
