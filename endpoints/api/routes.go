package api

import (
	"context"
	"sync"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	hubmcp "github.com/fsuhrau/automationhub/endpoints/mcp"
	"github.com/fsuhrau/automationhub/events"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/tester"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service struct {
	logger          *logrus.Entry
	db              *gorm.DB
	devicesManager  manager.Devices
	sessionsManager manager.Sessions
	nodeManager     manager.Nodes
	nodeUrl         string
	sseBroker       *sse.Broker
	cfg             config.Service

	runners      map[string]tester.Interface
	runnersMutex sync.Mutex

	control    *hubcontrol.Service
	runTracker *runs.Tracker
	mcp        *hubmcp.Handler
}

func New(logger *logrus.Logger, db *gorm.DB, nodeUrl string, dm manager.Devices, sm manager.Sessions, config config.Service, nodeManager manager.Nodes) *Service {
	svc := &Service{
		logger:          logger.WithField("Service", "Api"),
		nodeUrl:         nodeUrl,
		db:              db,
		devicesManager:  dm,
		sessionsManager: sm,
		nodeManager:     nodeManager,
		sseBroker:       sse.NewBroker(),
		cfg:             config,
		runners:         make(map[string]tester.Interface),
		control:         hubcontrol.New(dm),
	}
	svc.control.OnChange(svc.onDeviceControlChanged)
	svc.runTracker = runs.NewTracker(svc.onRunFinished)
	events.TestRunFinished.Register(svc.runTracker)
	svc.mcp = hubmcp.NewHandler(hubmcp.Deps{
		DB:      db,
		Backend: svc,
		Devices: dm,
		Control: svc.control,
		Runs:    svc.runTracker,
		NodeURL: nodeUrl,
	})
	return svc
}

func (s *Service) RegisterRoutes(r *gin.Engine, auth *gin.RouterGroup) error {
	api := auth.Group("/api")
	api.Use(httpx.WithLogger(s.logger))

	s.initSSE(api)
	go s.control.Run(context.Background())

	s.registerStatsRoutes(api)
	s.registerProfileRoutes(api)
	s.registerSettingsRoutes(api)
	s.registerProjectRoutes(api)
	s.registerDataRoutes(api)

	projectApi := api.Group("/:project_id")
	projectApi.Use(s.ResolveProject)
	{
		s.registerDeviceRoutes(projectApi)
		s.registerDeviceProtocolRoutes(projectApi)
		s.registerOverviewRoutes(projectApi)
		s.registerDeviceControlRoutes(projectApi)
		s.registerMCPRoutes(projectApi)
		s.registerAppRoutes(projectApi)

		appApi := projectApi.Group("/app/:app_id")
		appApi.Use(s.ResolveApp)
		{
			s.registerAppDetailRoutes(appApi)
			s.registerBundleRoutes(appApi)
			s.registerTestRoutes(appApi)
			s.registerTestRunRoutes(appApi)
			s.registerProtocolRoutes(appApi)
			s.registerBaselineRoutes(appApi)
		}
	}

	return nil
}
