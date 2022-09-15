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

	api.POST("/:project_id/app", s.createApp)
	api.GET("/:project_id/apps", s.getApps)
	// api.GET("/:project_id/app/:app_id", s.getApp)
	api.PUT("/:project_id/app/:app_id/binary/:binary_id", s.updateBinary)
	api.PUT("/:project_id/app/:app_id/binaries", s.getBinaries)
	api.DELETE("/:project_id/app/:app_id/binary/:binary_id", s.deleteBinary)
	api.POST("/:project_id/app/:app_id/upload", s.uploadBinary)

	api.POST("/:project_id/app/:app_id/test", s.newTest)
	api.GET("/data/test/data/:name", s.getData)
	api.GET("/:project_id/app/:app_id/test/:test_id", s.getTest)
	api.PUT("/:project_id/app/:app_id/test/:test_id", s.updateTest)
	api.POST("/:project_id/app/:app_id/test/:test_id/run", s.runTest)
	api.GET("/:project_id/app/:app_id/test/:test_id/runs", s.getTestRuns)
	api.GET("/:project_id/app/:app_id/test/:test_id/runs/last", s.getLastTestRun)
	api.GET("/:project_id/app/:app_id/test/:test_id/run/:run_id", s.getTestRun)
	api.GET("/:project_id/app/:app_id/test/:test_id/run/:run_id/:protocol_id", s.getTestRunProtocol)
	api.GET("/:project_id/app/:app_id/tests", s.getTests)

	api.GET("/:project_id/device/:device_id", s.getDevice)
	api.DELETE("/:project_id/device/:device_id", s.deleteDevice)
	api.PUT("/:project_id/device/:device_id", s.updateDevice)
	api.POST("/:project_id/device/:device_id/tests", s.deviceRunTests)
	api.GET("/:project_id/devices", s.getDevices)
	api.GET("/:project_id/device/connect", func(c *gin.Context) {
		s.socketHandler(c, c.Writer, c.Request)
	})

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
