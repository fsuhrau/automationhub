package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/authentication/github"
	"github.com/fsuhrau/automationhub/authentication/oauth2"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net"
	"net/http"
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
		if s.cfg.Auth.Token != nil && len(s.cfg.Auth.Token.AuthToken) > 0 {
			if token != s.cfg.Auth.Token.AuthToken {
				context.AbortWithStatus(http.StatusUnauthorized)
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

	api.POST("/app", s.createApp)
	api.GET("/app/:app_id", s.getApp)
	api.PUT("/app/:app_id", s.updateApp)
	api.DELETE("/app/:app_id", s.deleteApp)
	api.GET("/app/:app_id/functions", s.getAppFunctions)
	api.POST("/app/upload", s.uploadApp)
	api.GET("/apps", s.getApps)

	api.GET("/device/:device_id", s.getDevice)
	api.DELETE("/device/:device_id", s.deleteDevice)
	api.PUT("/device/:device_id", s.updateDevice)
	api.POST("/device/:device_id/tests", s.deviceRunTests)
	api.GET("/devices", s.getDevices)
	api.GET("/device/connect", func(c *gin.Context) {
		s.socketHandler(c, c.Writer, c.Request)
	})

	api.POST("/test", s.newTest)
	api.GET("/data/test/data/:name", s.getData)
	api.GET("/test/:test_id", s.getTest)
	api.PUT("/test/:test_id", s.updateTest)
	api.POST("/test/:test_id/run", s.runTest)
	api.GET("/test/:test_id/runs", s.getTestRuns)
	api.GET("/test/:test_id/runs/last", s.getLastTestRun)
	api.GET("/test/:test_id/run/:run_id", s.getTestRun)
	api.GET("/test/:test_id/run/:run_id/:protocol_id", s.getTestRunProtocol)
	api.GET("/tests", s.getTests)

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
