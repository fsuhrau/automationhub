package api

import (
	"context"
	"github.com/fsuhrau/automationhub/config"
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
	"sync"
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

	s.initSSE(api)
	go s.control.Run(context.Background())

	api.GET("/stats", s.getStats)
	s.registerProfile(api)

	s.registerSettings(api)

	api.GET("/projects", s.getProjects)
	api.POST("/project", s.createProject)
	api.DELETE("/project/:project_id", s.deleteProject)
	api.PUT("/project/:project_id", s.updateProject)

	api.GET("/data/:name", s.getData)

	projectApi := api.Group("/:project_id")
	projectApi.Use(s.ResolveProject)
	{
		projectApi.GET("/device/:device_id", s.WithProject(s.getDevice))
		projectApi.DELETE("/device/:device_id", s.WithProject(s.deleteDevice))
		projectApi.POST("/device/:device_id/unlock", s.WithProject(s.unlockDevice))
		projectApi.PUT("/device/:device_id", s.WithProject(s.updateDevice))
		projectApi.POST("/device/:device_id/tests", s.WithProject(s.deviceRunTests))
		projectApi.GET("/device/:device_id/protocols", s.WithProject(s.getDeviceProtocols))
		projectApi.GET("/overview", s.WithProject(s.getOverview))
		projectApi.GET("/devices", s.WithProject(s.getDevices))
		s.registerDeviceControl(projectApi)
		// MCP (streamable HTTP): POST for messages, GET for the event stream, DELETE ends a session
		projectApi.POST("/mcp", s.WithProject(s.serveMCP))
		projectApi.GET("/mcp", s.WithProject(s.serveMCP))
		projectApi.DELETE("/mcp", s.WithProject(s.serveMCP))
		projectApi.PUT("/mcp/uploads/:upload_id", s.WithProject(s.serveMCPUpload))

		projectApi.POST("/app", s.WithProject(s.createApp))
		projectApi.GET("/apps", s.WithProject(s.getApps))

		appApi := projectApi.Group("/app/:app_id")
		appApi.Use(s.ResolveApp)
		{
			appApi.PUT("/", s.WithApp(s.updateApp))
			appApi.POST("/parameter", s.WithApp(s.addAppParameter))
			appApi.PUT("/parameter/:parameter_id", s.WithApp(s.updateAppParameter))
			appApi.DELETE("/parameter/:parameter_id", s.WithApp(s.removeAppParameter))
			appApi.GET("/bundles", s.WithApp(s.getBinaries))
			appApi.PUT("/bundle/:binary_id", s.WithApp(s.updateBinary))
			appApi.DELETE("/bundle/:binary_id", s.WithApp(s.deleteBinary))
			appApi.POST("/upload", s.WithApp(s.uploadBinary))
			appApi.POST("/test", s.WithApp(s.newTest))
			appApi.GET("/test/:test_id", s.WithApp(s.getTest))
			appApi.PUT("/test/:test_id", s.WithApp(s.updateTest))
			appApi.POST("/test/:test_id/run", s.WithApp(s.runTest))
			appApi.GET("/test/:test_id/runs", s.WithApp(s.getTestRuns))
			appApi.GET("/test/:test_id/runs/last", s.WithApp(s.getLastTestRun))
			appApi.GET("/test/:test_id/run/:run_id", s.WithApp(s.getTestRun))
			appApi.POST("/test/:test_id/run/:run_id/cancel", s.WithApp(s.cancelTestRun))
			appApi.GET("/test/:test_id/run/:run_id/:protocol_id", s.WithApp(s.getTestRunProtocol))
			appApi.GET("/tests", s.WithApp(s.getTests))
		}
	}

	return nil
}

// registerSettings adds the project settings. They hand out and revoke
// tokens, so only a browser session may use them: a project token must not
// mint a node token, or tokens for other projects.
func (s *Service) registerSettings(api *gin.RouterGroup) {
	settings := api.Group("/:project_id/settings", s.requireSession)
	settings.GET("/users", s.getUsers)

	settings.POST("/access_token", s.createAccessToken)
	settings.GET("/access_tokens", s.getAccessTokens)
	settings.DELETE("/access_token/:token_id", s.deleteAccessToken)

	settings.POST("/node", s.createNode)
	settings.GET("/nodes", s.getNodes)
	settings.DELETE("/nodes/:node_id", s.deleteNode)
	settings.GET("/nodes/:node_id", s.getNodeStatus)
	settings.POST("/nodes/:node_id/token", s.issueNodeTokenHandler)
}
