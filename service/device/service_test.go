package device_test

import (
	"context"
	"math"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
	svc "github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

// env is a service on fakes.
type env struct {
	svc      *svc.Service
	repo     *fakeRepo
	pool     *testkit.FakeDeviceManager
	leases   *fakeLeases
	executor *fakeExecutor
	project  *models.Project
}

func newEnv(devices ...models.Device) *env {
	e := &env{
		repo:     &fakeRepo{devices: devices},
		pool:     &testkit.FakeDeviceManager{},
		leases:   &fakeLeases{holders: map[string]string{}},
		executor: &fakeExecutor{},
		project:  &models.Project{Model: models.Model{ID: 7}},
	}
	e.svc = svc.New(svc.Deps{
		Repo:        e.repo,
		Devices:     e.pool,
		Leases:      e.leases,
		NewExecutor: func() svc.Executor { return e.executor },
		StoreDir:    "/store",
	})
	return e
}

func stored(id uint, identifier string) models.Device {
	return models.Device{Model: models.Model{ID: id}, DeviceIdentifier: identifier}
}

func assertKind(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	if got := apperr.KindOf(err); got != want {
		t.Fatalf("error kind = %v (%v), want %v", got, err, want)
	}
}

type liveCase struct {
	name       string
	stored     models.Device
	live       device.Device
	holder     string
	wantStatus device.State
	wantLocked bool
	wantBy     string
}

func liveCases() []liveCase {
	online := testkit.NewFakeDevice("dev-1")
	online.State = device.StateBooted
	_ = online.Lock()
	online.SetConnection(&device.Connection{})
	node := stored(1, "dev-1")
	node.NodeID = 3
	return []liveCase{
		{"offline", stored(1, "dev-1"), nil, "", device.StateUnknown, false, ""},
		{"node device offline", node, nil, "", device.StateNodeDisconnected, false, ""},
		{"manager says booted", stored(1, "dev-1"), online, "", device.StateBooted, true, ""},
		{"lease held", stored(1, "dev-1"), online, "alice", device.StateBooted, true, "alice"},
	}
}

func (tc liveCase) check(t *testing.T, got models.Device) {
	t.Helper()
	if got.Status != tc.wantStatus || got.IsLocked != tc.wantLocked || got.LockedBy != tc.wantBy {
		t.Fatalf("got status %v locked %v by %q", got.Status, got.IsLocked, got.LockedBy)
	}
	if (got.Dev != nil) != (tc.live != nil) || (got.Connection != nil) != (tc.live != nil) {
		t.Fatalf("dev %v connection %v", got.Dev, got.Connection)
	}
}

// The list and the detail view merge the pool and the leases the same way.
func TestLiveStatusMerge(t *testing.T) {
	for _, tc := range liveCases() {
		for _, via := range []string{"list", "get"} {
			t.Run(tc.name+" via "+via, func(t *testing.T) {
				e := newEnv(tc.stored)
				if tc.live != nil {
					e.pool.Add("android_device", tc.live)
				}
				if tc.holder != "" {
					e.leases.holders["dev-1"] = tc.holder
				}
				tc.check(t, e.fetch(t, via))
			})
		}
	}
}

// fetch loads device 1 through List or Get.
func (e *env) fetch(t *testing.T, via string) models.Device {
	t.Helper()
	if via == "get" {
		d, err := e.svc.Get(context.Background(), e.project, "1")
		if err != nil {
			t.Fatal(err)
		}
		return *d
	}
	list, err := e.svc.List(context.Background(), e.project, nil)
	if err != nil || len(list) != 1 {
		t.Fatalf("list %v %v", list, err)
	}
	return list[0]
}

func TestListPassesProjectAndPlatform(t *testing.T) {
	e := newEnv(stored(1, "dev-1"))
	platform := models.PlatformTypeAndroid

	if _, err := e.svc.List(context.Background(), e.project, &platform); err != nil {
		t.Fatal(err)
	}
	if e.repo.projects[0] != e.project || e.repo.platform == nil || *e.repo.platform != platform {
		t.Fatalf("repo got project %v platform %v", e.repo.projects, e.repo.platform)
	}
}

func TestReadErrorsKeepTheirKind(t *testing.T) {
	e := newEnv(stored(1, "dev-1"))
	e.repo.fail = map[string]error{"List": apperr.Internal("db down"), "FindDetail": apperr.Internal("db down")}

	_, err := e.svc.List(context.Background(), e.project, nil)
	assertKind(t, err, apperr.KindInternal)
	_, err = e.svc.Get(context.Background(), e.project, "1")
	assertKind(t, err, apperr.KindInternal)
}

func TestGetOfAnotherProjectsDeviceIsNotFound(t *testing.T) {
	e := newEnv(stored(1, "dev-1"))

	_, err := e.svc.Get(context.Background(), e.project, "2")

	assertKind(t, err, apperr.KindNotFound)
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name string
		id   string
		fail map[string]error
		want apperr.Kind
	}{
		{"deletes", "1", nil, apperr.KindNone},
		{"unknown", "2", nil, apperr.KindNotFound},
		{"storage fails", "1", map[string]error{"Delete": apperr.Internal("x")}, apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(stored(1, "dev-1"))
			e.repo.fail = tc.fail
			assertKind(t, e.svc.Delete(context.Background(), e.project, tc.id), tc.want)
		})
	}
}

type unlockCase struct {
	name      string
	connected bool
	locked    bool
	unlockErr error
	want      apperr.Kind
}

func (tc unlockCase) run(t *testing.T) {
	e := newEnv(stored(1, "dev-1"))
	dev := testkit.NewFakeDevice("dev-1")
	if tc.locked {
		_ = dev.Lock()
	}
	dev.FailWith("Unlock", tc.unlockErr)
	if tc.connected {
		e.pool.Add("android_device", dev)
	}

	d, err := e.svc.Unlock(context.Background(), e.project, "1")

	assertKind(t, err, tc.want)
	if (d != nil) != (tc.want == apperr.KindNone) {
		t.Fatalf("device returned = %v", d)
	}
	if released := len(e.leases.released) == 1; released != tc.connected {
		t.Fatalf("lease released = %v, want %v", released, tc.connected)
	}
	if stillLocked := dev.IsLocked(); stillLocked != (tc.locked && tc.unlockErr != nil) {
		t.Fatalf("device locked = %v", stillLocked)
	}
}

func TestUnlock(t *testing.T) {
	for _, tc := range []unlockCase{
		{"not connected", false, false, nil, apperr.KindNotFound},
		{"connected and locked", true, true, nil, apperr.KindNone},
		{"connected, not locked", true, false, nil, apperr.KindNone},
		{"unlock fails", true, true, errBoom, apperr.KindConflict},
	} {
		t.Run(tc.name, tc.run)
	}
}

func TestUnlockUnknownDevice(t *testing.T) {
	e := newEnv()
	_, err := e.svc.Unlock(context.Background(), e.project, "1")
	assertKind(t, err, apperr.KindNotFound)
}

func TestProtocols(t *testing.T) {
	e := newEnv(stored(1, "dev-1"))
	e.repo.protocols = []models.TestProtocol{{AvgFPS: math.NaN()}, {AvgMEM: math.NaN()}}

	got, err := e.svc.Protocols(context.Background(), e.project, svc.ProtocolQuery{DeviceID: "1", Limit: 5})

	if err != nil || len(got) != 2 || got[0].AvgFPS != 0 || got[1].AvgMEM != 0 || e.repo.limit != 5 {
		t.Fatalf("got %+v err %v limit %d", got, err, e.repo.limit)
	}
}

func TestProtocolsErrors(t *testing.T) {
	e := newEnv(stored(1, "dev-1"))
	_, err := e.svc.Protocols(context.Background(), e.project, svc.ProtocolQuery{DeviceID: "2", Limit: 5})
	assertKind(t, err, apperr.KindNotFound)

	e.repo.fail = map[string]error{"Protocols": apperr.Internal("x")}
	_, err = e.svc.Protocols(context.Background(), e.project, svc.ProtocolQuery{DeviceID: "1", Limit: 5})
	assertKind(t, err, apperr.KindInternal)
}

func TestControlDevice(t *testing.T) {
	e := newEnv(stored(1, "dev-1"))
	_, err := e.svc.ControlDevice(context.Background(), e.project, "2")
	assertKind(t, err, apperr.KindNotFound)
	_, err = e.svc.ControlDevice(context.Background(), e.project, "1")
	assertKind(t, err, apperr.KindConflict)

	live := testkit.NewFakeDevice("dev-1")
	e.pool.Add("android_device", live)
	got, err := e.svc.ControlDevice(context.Background(), e.project, "1")
	if err != nil || got != device.Device(live) {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestControlApp(t *testing.T) {
	tests := []struct {
		name   string
		bundle *models.AppBinary
		fail   error
		want   apperr.Kind
	}{
		{"found", &models.AppBinary{Model: models.Model{ID: 4}, Identifier: "com.x", AppPath: "a.apk"}, nil, apperr.KindNone},
		{"archived", &models.AppBinary{ArchivedAt: ptrNow()}, nil, apperr.KindInvalid},
		{"missing", nil, apperr.NotFound("app bundle 4 not found"), apperr.KindNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv()
			e.repo.bundle, e.repo.fail = tc.bundle, map[string]error{"Bundle": tc.fail}

			params, err := e.svc.ControlApp(context.Background(), e.project, 4)

			assertKind(t, err, tc.want)
			if tc.want == apperr.KindNone && (params.Identifier != "com.x" || params.App.AppPath != "/store/a.apk") {
				t.Fatalf("params %+v", params)
			}
		})
	}
}
