package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/hub/action"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/internal/testkit"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	devicesvc "github.com/fsuhrau/automationhub/service/device"
	runsvc "github.com/fsuhrau/automationhub/service/run"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/appstore"
	"github.com/fsuhrau/automationhub/storage/devicestore"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/runstore"
	"github.com/fsuhrau/automationhub/storage/teststore"
	"gorm.io/gorm"
)

// goldenWorld is the seeded project the golden tests call every tool on: the
// real services on a testkit database, a fake device and a fake runner.
type goldenWorld struct {
	db       *gorm.DB
	project  *models.Project
	pool     *appPool
	dev      *testkit.FakeDevice
	control  *hubcontrol.Service
	tracker  *runs.Tracker
	runs     *runsvc.Service
	apps     *appsvc.Service
	tests    *testsvc.Service
	files    *httptest.Server
	devices  *devicesvc.Service
	dataDir  string
	storeDir string
}

// appPool is the pool of devices whose app, if one is connected, answers the
// scene graph and screenshot requests with the scene and screenshot it holds.
type appPool struct {
	testkit.FakeDeviceManager
	scene      string
	screenshot []byte
}

func (p *appPool) SendAction(dev device.Device, a action.Interface) error {
	resp := &appproto.Response{ActionType: a.GetActionType(), Success: true}
	switch a.GetActionType() {
	case appproto.ActionType_GetSceneGraph:
		resp.Payload.Screenshot = &appproto.Screenshot{ContentType: appproto.ContentType_Xml, Sceengraph: []byte(p.scene)}
	case appproto.ActionType_GetScreenshot:
		resp.Payload.Screenshot = &appproto.Screenshot{Screenshot: p.screenshot}
	}
	go func() {
		for _, h := range dev.(*testkit.FakeDevice).ActionHandlers() {
			h.OnActionResponse(dev, resp)
		}
	}()
	return nil
}

// goldenAnalyser reads a file as the build "Client 1.58.0" whose hash is the
// hash of its content, so equal uploads are the same build.
type goldenAnalyser struct{}

func (goldenAnalyser) Analyse(path string) (*hubapp.Parameter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	return &hubapp.Parameter{
		Platform: "android", Identifier: "com.x.android", Name: "Client", Version: "1.58.0",
		App: &hubapp.AppParams{Hash: hex.EncodeToString(sum[:4]), Size: len(data)},
	}, nil
}

func (goldenAnalyser) Read(string) (*hubapp.Parameter, error) { return &hubapp.Parameter{}, nil }

func (goldenAnalyser) Icon(string, string) ([]byte, string, error) { return nil, "", nil }

// goldenRunner starts the run the fake runner stores for its test.
type goldenRunner struct {
	db   *gorm.DB
	test models.Test
}

func (r *goldenRunner) Initialize(test models.Test, _ map[string]string) error {
	r.test = test
	return nil
}

func (r *goldenRunner) Run(_ []models.Device, bundle *models.AppBinary, startURL string) (*models.TestRun, error) {
	started := &models.TestRun{TestID: r.test.ID, SessionID: "session-1", StartURL: startURL}
	if bundle != nil {
		started.AppBinaryID = bundle.ID
	}
	return started, r.db.Create(started).Error
}

func (r *goldenRunner) Cancel(string) error   { return nil }
func (r *goldenRunner) LockedDevices() []uint { return nil }

// goldenRunners is the in-memory book of active runners.
type goldenRunners map[uint]runsvc.Runner

func (b goldenRunners) Add(id uint, r runsvc.Runner)      { b[id] = r }
func (b goldenRunners) Get(id uint) (runsvc.Runner, bool) { r, ok := b[id]; return r, ok }
func (b goldenRunners) Remove(id uint)                    { delete(b, id) }

// newBareWorld is the world without a database: the pool with its one live
// device, the live control, the tracker, the folders and the file server.
func newBareWorld(t *testing.T) *goldenWorld {
	t.Helper()
	w := &goldenWorld{pool: &appPool{}, tracker: runs.NewTracker(nil)}
	w.project = &models.Project{Model: models.Model{ID: 1}, Identifier: "p1", Name: "P1"}
	w.dataDir, w.storeDir = t.TempDir(), t.TempDir()
	w.control = hubcontrol.New(w.pool)
	w.dev = testkit.NewFakeDevice("dev-1")
	w.dev.State, w.dev.Parameter = device.StateBooted, map[string]string{"abi": "arm64"}
	w.dev.PNG, w.dev.Width, w.dev.Height = testPNG(t, 2000, 1000), 2000, 1000
	w.dev.LogLines = []string{"line one", "line two"}
	w.pool.Add("fake", w.dev)
	w.files = httptest.NewServer(http.HandlerFunc(serveBuild))
	t.Cleanup(w.files.Close)
	return w
}

// newGoldenWorld is the bare world with the services on a seeded database.
func newGoldenWorld(t *testing.T) *goldenWorld {
	t.Helper()
	w := newBareWorld(t)
	w.db = testkit.DB(t)
	w.devices = devicesvc.New(devicesvc.Deps{Repo: devicestore.New(w.db), Devices: w.pool, Leases: w.control})
	w.tests = testsvc.New(testsvc.Deps{Repo: teststore.New(w.db)})
	w.runs = runsvc.New(runsvc.Deps{
		Repo: runstore.New(w.db), Pool: w.pool, Runners: goldenRunners{}, Summarizer: w.tests,
		NewRunner: func(models.TestType, *models.Project, uint) (runsvc.Runner, error) {
			return &goldenRunner{db: w.db}, nil
		},
		RunState: func(id uint) runs.RunState { s, _ := w.tracker.Status(id); return s },
		Started:  w.tracker.Started,
	})
	w.apps = appsvc.New(appsvc.Deps{
		Repo: appstore.New(w.db), Files: appsvc.NewOSFiles(w.storeDir), Analyser: goldenAnalyser{},
		Running: func(uint) bool { return false },
		Summarize: func(r []models.TestRun) ([]appsvc.RunSummary, error) {
			return w.tests.Summaries(context.Background(), r)
		},
	})
	w.seed(t)
	return w
}

// serveBuild serves builds for upload_bundle: /ok.apk and /dup.apk are the
// same file, /missing.apk is a 404.
func serveBuild(w http.ResponseWriter, r *http.Request) {
	switch filepath.Base(r.URL.Path) {
	case "ok.apk", "dup.apk":
		_, _ = w.Write([]byte("build content"))
	default:
		http.NotFound(w, r)
	}
}

func at(h, m, s int) time.Time { return time.Date(2026, 1, 1, h, m, s, 0, time.UTC) }

func withDevice(set func(*models.Device)) testkit.Option[models.Device] { return set }

// seed stores two projects with apps, bundles, tests, runs, protocols and
// devices; the ids are the order they are created in.
func (w *goldenWorld) seed(t *testing.T) {
	t.Helper()
	db := w.db
	db.Create(w.project)
	other := testkit.Project(db, func(p *models.Project) { p.Identifier = "p2"; p.Name = "P2" })
	node := &models.Node{Identifier: "n1", Name: "Lab"}
	db.Create(node)

	testkit.Device(db, withDevice(func(d *models.Device) {
		d.Name, d.DeviceIdentifier, d.PlatformType, d.OS, d.OSVersion, d.NodeID = "Pixel", "dev-1", models.PlatformTypeAndroid, "Android", "14", node.ID
	}))
	testkit.Device(db, withDevice(func(d *models.Device) {
		d.Name, d.Alias, d.DeviceIdentifier, d.PlatformType, d.NodeID = "iPhone", "Pocket", "dev-2", models.PlatformTypeiOS, node.ID
	}))
	testkit.Device(db, withDevice(func(d *models.Device) {
		d.Name, d.DeviceIdentifier, d.PlatformType = "Box", "dev-3", models.PlatformTypeLinux
	}))
	w.seedApps(other.ID)
	w.seedRuns(t)
}

func (w *goldenWorld) seedApps(otherProject uint) {
	db := w.db
	android := testkit.App(db, func(a *models.App) { a.ProjectID, a.Name, a.Identifier = w.project.ID, "Android", "com.x.android" })
	web := testkit.App(db, func(a *models.App) {
		a.ProjectID, a.Name, a.Identifier, a.Platform = w.project.ID, "Web", "web", models.PlatformTypeWeb
	})
	foreign := testkit.App(db, func(a *models.App) { a.ProjectID, a.Name, a.Identifier = otherProject, "Foreign", "com.x.foreign" })

	archived := at(9, 0, 0)
	db.Create(&models.AppBinary{AppID: android.ID, Name: "Client", Version: "1.0", Identifier: "com.x.android", Platform: "android",
		Size: 1234, Tags: "release", AppPath: "client_h1.apk", Hash: "h1"})
	db.Create(&models.AppBinary{AppID: android.ID, Name: "Client", Version: "0.9", Platform: "android", Size: 99, AppPath: "client_h0.apk", Hash: "h0", ArchivedAt: &archived})
	db.Create(&models.AppBinary{AppID: foreign.ID, Name: "Foreign", Version: "1.0", Platform: "android", AppPath: "f.apk", Hash: "hf"})

	scenario := func(app uint, name string, kind models.TestType, all bool, devices ...uint) {
		cfg := models.TestConfig{Type: kind, AllDevices: all}
		for _, id := range devices {
			cfg.Devices = append(cfg.Devices, models.TestConfigDevice{DeviceID: id})
		}
		db.Create(&models.Test{AppID: app, Name: name, TestConfig: cfg})
	}
	scenario(android.ID, "Smoke", models.TestTypeScenario, false, 1)
	scenario(android.ID, "Unity", models.TestTypeUnity, true)
	scenario(foreign.ID, "Foreign test", models.TestTypeScenario, false, 1)
	scenario(web.ID, "Web smoke", models.TestTypeScenario, false, 1)
}

func ptrTime(v time.Time) *time.Time { return &v }

func (w *goldenWorld) seedRuns(t *testing.T) {
	t.Helper()
	db := w.db
	r1 := &models.TestRun{TestID: 1, AppBinaryID: 1, SessionID: "s1", Parameter: "Level=1\nProxy=x"}
	db.Create(r1)
	rForeign := &models.TestRun{TestID: 3, SessionID: "s3"}
	db.Create(rForeign)
	dev := uint(1)
	p1 := &models.TestProtocol{TestRunID: r1.ID, DeviceID: &dev, TestName: "Suite/A", StartedAt: at(10, 0, 0), EndedAt: ptrTime(at(10, 1, 35)),
		TestResult: models.TestResultSuccess, AvgFPS: 59.5, AvgMEM: 120}
	db.Create(p1)
	db.Create(&models.TestProtocol{TestRunID: r1.ID, ParentTestProtocolID: &p1.ID, DeviceID: &dev, TestName: "Suite/A/Child",
		StartedAt: at(10, 0, 5), EndedAt: ptrTime(at(10, 0, 9)), TestResult: models.TestResultFailed})
	db.Create(&models.TestProtocol{TestRunID: r1.ID, TestName: "Suite/B", StartedAt: at(10, 2, 0)})
	foreignP := &models.TestProtocol{TestRunID: rForeign.ID, TestName: "Foreign/A", StartedAt: at(10, 0, 0)}
	db.Create(foreignP)

	db.Create(&models.ProtocolKey{TestProtocolID: p1.ID, Key: "player_id", Value: "1234"})
	entry := func(pid uint, level, source, msg, data string, runtime float64, sec int) {
		db.Create(&models.ProtocolEntry{TestProtocolID: pid, Timestamp: at(10, 0, sec), Level: level, Source: source, Message: msg, Data: data, Runtime: runtime})
	}
	entry(p1.ID, "info", "log", "starting", "", 0, 1)
	entry(p1.ID, "info", "screen", "login", "first.png", 1.5, 2)
	entry(p1.ID, "error", "log", "boom 50% failed", "", 2, 3)
	entry(p1.ID, "exception", "log", "NullReference", "trace.json", 3, 4)
	entry(p1.ID, "info", "screen", "", "last.png", 9, 5)
	entry(p1.ID, "info", "screen", "gone", "gone.png", 10, 6)
	entry(p1.ID, "info", "screen", "broken", "bad.png", 11, 7)
	entry(foreignP.ID, "info", "screen", "", "foreign.png", 1, 1)
	db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p1.ID, Checkpoint: "start", FPS: 60, MEM: 100, CPU: 10})
	db.Create(&models.ProtocolPerformanceEntry{TestProtocolID: p1.ID, Checkpoint: "end", FPS: 50, MEM: 140, CPU: 30})
	for _, l := range [][2]string{{"info", "run started"}, {"error", "device lost"}, {"info", "run ended"}} {
		db.Create(&models.TestRunLogEntry{TestRunID: r1.ID, Level: l[0], Log: l[1]})
	}

	write := func(name string, data []byte) {
		if err := os.WriteFile(filepath.Join(w.dataDir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("first.png", testPNG(t, 400, 800))
	write("last.png", testPNG(t, 2000, 1000))
	write("foreign.png", testPNG(t, 10, 10))
	write("bad.png", []byte("not an image"))
}
