package api

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/manager"
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
}

func New(logger *logrus.Logger, db *gorm.DB, nodeUrl string, dm manager.Devices, sm manager.Sessions, config config.Service, nodeManager manager.Nodes) *Service {
	return &Service{
		logger:          logger.WithField("Service", "Api"),
		nodeUrl:         nodeUrl,
		db:              db,
		devicesManager:  dm,
		sessionsManager: sm,
		nodeManager:     nodeManager,
		sseBroker:       sse.NewBroker(),
		cfg:             config,
		runners:         make(map[string]tester.Interface),
	}
}

func (s *Service) RegisterRoutes(r *gin.Engine, auth *gin.RouterGroup) error {
	api := auth.Group("/api")

	s.initSSE(api)

	api.GET("/stats", s.getStats)

	api.GET("/:project_id/settings/users", s.getUsers)

	api.POST("/:project_id/settings/access_token", s.createAccessToken)
	api.GET("/:project_id/settings/access_tokens", s.getAccessTokens)
	api.DELETE("/:project_id/settings/access_token/:token_id", s.deleteAccessToken)

	api.POST("/:project_id/settings/node", s.createNode)
	api.GET("/:project_id/settings/nodes", s.getNodes)
	api.DELETE("/:project_id/settings/nodes/:node_id", s.deleteNode)
	api.GET("/:project_id/settings/nodes/:node_id", s.getNodeStatus)

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
		projectApi.GET("/devices", s.WithProject(s.getDevices))

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
