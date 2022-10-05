package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/authentication/github"
	"github.com/fsuhrau/automationhub/authentication/oauth2"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net"
	"net/http"
	"time"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Service struct {
	logger          *logrus.Entry
	db              *gorm.DB
	devicesManager  manager.Devices
	sessionsManager manager.Sessions
	hostIP          net.IP
	sseBroker       *sse.Broker
	cfg             config.Service
}

func New(logger *logrus.Logger, db *gorm.DB, ip net.IP, dm manager.Devices, sm manager.Sessions, config config.Service) *Service {
	return &Service{
		logger:          logger.WithField("Service", "Api"),
		hostIP:          ip,
		db:              db,
		devicesManager:  dm,
		sessionsManager: sm,
		sseBroker:       sse.NewBroker(),
		cfg:             config,
	}
}

func (s *Service) handleSessionAccess(fallbackFunc gin.HandlerFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		token := context.GetHeader("X-Auth-Token")
		if len(token) == 0 && fallbackFunc != nil {
			fallbackFunc(context)
			return
		}

		// TODO: handle token authentication via web interface
		if s.cfg.Auth.Token != nil {
			var accessToken models.AccessToken
			if err := s.db.First(&accessToken, "token = ?", token).Error; err != nil {
				context.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			if accessToken.Token != token {
				context.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			if accessToken.ExpiresAt != nil && accessToken.ExpiresAt.Before(time.Now()) {
				context.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}
	}
}

func (s *Service) RegisterRoutes(r *gin.Engine) error {
	api := r.Group("/api")

	if s.cfg.Auth.AuthenticationRequired() {
		if s.cfg.Auth.OAuth2 != nil {
			api.Use(s.handleSessionAccess(oauth2.Auth()))
		} else if s.cfg.Auth.Github != nil {
			api.Use(s.handleSessionAccess(github.Auth()))
		} else {
			api.Use(s.handleSessionAccess(nil))
		}
	}

	s.initSSE(api)

	api.GET("/stats", s.getStats)

	api.POST("/:project_id/settings/access_token", s.createAccessToken)
	api.GET("/:project_id/settings/access_tokens", s.getAccessTokens)
	api.DELETE("/:project_id/settings/access_token/:token_id", s.deleteAccessToken)

	api.GET("/projects", s.getProjects)
	api.POST("/project", s.createProject)
	api.DELETE("/project/:project_id", s.deleteProject)
	api.PUT("/project/:project_id", s.updateProject)

	api.GET("/data/test/data/:name", s.getData)

	projectApi := api.Group("/:project_id")
	projectApi.Use(s.ResolveProject)
	{
		projectApi.GET("/device/:device_id", s.WithProject(s.getDevice))
		projectApi.DELETE("/device/:device_id", s.WithProject(s.deleteDevice))
		projectApi.PUT("/device/:device_id", s.WithProject(s.updateDevice))
		projectApi.POST("/device/:device_id/tests", s.WithProject(s.deviceRunTests))
		projectApi.GET("/devices", s.WithProject(s.getDevices))

		projectApi.GET("/device/connect", func(c *gin.Context) {
			s.socketHandler(c, c.Writer, c.Request)
		})

		projectApi.POST("/app", s.WithProject(s.createApp))
		projectApi.GET("/apps", s.WithProject(s.getApps))

		appApi := projectApi.Group("/app/:app_id")
		appApi.Use(s.ResolveApp)
		{
			appApi.PUT("/", s.WithApp(s.updateApp))
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
			appApi.GET("/test/:test_id/run/:run_id/:protocol_id", s.WithApp(s.getTestRunProtocol))
			appApi.GET("/tests", s.WithApp(s.getTests))
		}
	}

	return nil
}

func (s *Service) socketHandler(c *gin.Context, w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Fail to get register request: %+v", err)
		return
	}
	s.registerDevices(msg, conn, c)
}
