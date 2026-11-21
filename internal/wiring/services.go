// Package wiring is the composition root of the hub: it builds the database
// connection, stores, services, managers and endpoints of a master or a node
// exactly once and hands each of them only what it needs. cmd/ parses flags
// and config and calls BuildMaster or BuildNode.
package wiring

import (
	"context"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/events"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/service/access"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	devicesvc "github.com/fsuhrau/automationhub/service/device"
	overviewsvc "github.com/fsuhrau/automationhub/service/overview"
	"github.com/fsuhrau/automationhub/service/project"
	protocolsvc "github.com/fsuhrau/automationhub/service/protocol"
	runsvc "github.com/fsuhrau/automationhub/service/run"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/accessstore"
	"github.com/fsuhrau/automationhub/storage/appstore"
	"github.com/fsuhrau/automationhub/storage/devicestore"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/overviewstore"
	"github.com/fsuhrau/automationhub/storage/projectstore"
	"github.com/fsuhrau/automationhub/storage/runstore"
	"github.com/fsuhrau/automationhub/storage/teststore"
	"github.com/fsuhrau/automationhub/tester/unity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Infra is what the services of a master are built on.
type Infra struct {
	DB      *gorm.DB
	Devices manager.Devices
	Nodes   manager.Nodes
	Config  config.Service
	Logger  *logrus.Logger
}

// services are the use cases of a master, each built once.
type services struct {
	projects  *project.Service
	access    *access.Service
	apps      *appsvc.Service
	tests     *testsvc.Service
	runs      *runsvc.Service
	protocols *protocolsvc.Service
	overview  *overviewsvc.Service
	devices   *devicesvc.Service
	// mcpDevices is the device service of MCP, which doesn't run tests.
	mcpDevices *devicesvc.Service
}

// runtime are the stateful parts the services share.
type runtime struct {
	control *hubcontrol.Service
	tracker *runs.Tracker
	runners *runsvc.Registry
	broker  *sse.Broker
}

// newRuntime builds the device control, the SSE broker and the run tracking;
// a finished run is forgotten by the registry of active runners.
func newRuntime(devices manager.Devices) *runtime {
	rt := &runtime{control: hubcontrol.New(devices), runners: runsvc.NewRegistry(), broker: sse.NewBroker()}
	rt.tracker = runs.NewTracker(rt.runners.Remove)
	events.TestRunFinished.Register(rt.tracker)
	return rt
}

// runState tells whether this master is running the run right now.
func (rt *runtime) runState(runID uint) runs.RunState {
	state, _ := rt.tracker.Status(runID)
	return state
}

func newServices(in Infra, rt *runtime) *services {
	s := &services{}
	s.tests = testsvc.New(testsvc.Deps{Repo: teststore.New(in.DB), RunState: rt.runState})
	s.runs = newRunService(in, rt, s.tests)
	s.projects = project.New(project.Deps{Repo: projectstore.New(in.DB)})
	s.access = access.New(access.Deps{Repo: accessstore.New(in.DB), Status: in.Nodes})
	s.apps = newAppService(in, rt, s.tests)
	s.protocols = protocolsvc.New(protocolsvc.Deps{Repo: runstore.New(in.DB), Scope: s.runs, DeviceName: devicesvc.DisplayName})
	s.overview = overviewsvc.New(overviewsvc.Deps{Repo: overviewstore.New(in.DB), Summarizer: s.tests, Pool: in.Devices})
	s.mcpDevices = devicesvc.New(deviceDeps(in, rt))
	s.devices = devicesvc.New(withExecutor(deviceDeps(in, rt), in.Devices))
	return s
}

func newAppService(in Infra, rt *runtime, tests *testsvc.Service) *appsvc.Service {
	return appsvc.New(appsvc.Deps{
		Repo:     appstore.New(in.DB),
		Files:    appsvc.NewOSFiles(in.Config.Paths.AppStore),
		Analyser: appsvc.FileAnalyser{},
		Running:  func(runID uint) bool { return rt.runState(runID) == runs.RunRunning },
		Summarize: func(testRuns []models.TestRun) ([]appsvc.RunSummary, error) {
			return tests.Summaries(context.Background(), testRuns)
		},
	})
}

func deviceDeps(in Infra, rt *runtime) devicesvc.Deps {
	return devicesvc.Deps{
		Repo:     devicestore.New(in.DB),
		Devices:  in.Devices,
		Leases:   rt.control,
		StoreDir: in.Config.Paths.AppStore,
	}
}

// withExecutor lets the device service run tests on the devices of the pool.
func withExecutor(deps devicesvc.Deps, pool manager.Devices) devicesvc.Deps {
	deps.NewExecutor = func() devicesvc.Executor { return unity.NewExecutor(pool, nil) }
	return deps
}
