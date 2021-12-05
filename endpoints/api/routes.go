package api

import (
	"fmt"
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

type ApiService struct {
	logger          *logrus.Entry
	db              *gorm.DB
	devicesManager  manager.Devices
	sessionsManager manager.Sessions
	hostIP          net.IP
	sseBroker       *sse.Broker
}

func New(logger *logrus.Logger, db *gorm.DB, ip net.IP, dm manager.Devices, sm manager.Sessions) *ApiService {
	return &ApiService{
		logger:          logger.WithField("Service", "Api"),
		hostIP:          ip,
		db:              db,
		devicesManager:  dm,
		sessionsManager: sm,
		sseBroker:       sse.NewBroker(),
	}
}

func HandleWithSession(f func(*Session, *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		s, _ := c.Get("session")
		var session *Session
		if s != nil {
			session = s.(*Session)
			// session.LastAccess = time.Now()
		}
		f(session, c)
	}
}

func (s *ApiService) RegisterRoutes(r *gin.Engine) error {
	api := r.Group("/api")

	s.initSSE(api)

	api.POST("/app", HandleWithSession(s.createApp))
	api.GET("/app/:app_id", HandleWithSession(s.getApp))
	api.PUT("/app/:app_id", HandleWithSession(s.updateApp))
	api.DELETE("/app/:app_id", HandleWithSession(s.deleteApp))
	api.GET("/app/:app_id/functions", HandleWithSession(s.getAppFunctions))
	api.POST("/app/upload", HandleWithSession(s.uploadApp))
	api.GET("/apps", HandleWithSession(s.getApps))

	api.GET("/device/:device_id", HandleWithSession(s.getDevice))
	api.DELETE("/device/:device_id", HandleWithSession(s.deleteDevice))
	api.POST("/device/:device_id/tests", HandleWithSession(s.deviceRunTests))
	api.GET("/devices", HandleWithSession(s.getDevices))
	api.GET("/device/connect", func(c *gin.Context) {
		s.socketHandler(c, c.Writer, c.Request)
	})

	api.POST("/test", HandleWithSession(s.newTest))
	api.GET("/data/test/data/:name", HandleWithSession(s.getData))
	api.GET("/test/:test_id", HandleWithSession(s.getTest))
	api.PUT("/test/:test_id", HandleWithSession(s.updateTest))
	api.POST("/test/:test_id/run", s.runTest)
	api.GET("/test/:test_id/runs", HandleWithSession(s.getTestRuns))
	api.GET("/test/:test_id/runs/last", HandleWithSession(s.getLastTestRun))
	api.GET("/test/:test_id/run/:run_id", HandleWithSession(s.getTestRun))
	api.GET("/test/:test_id/run/:run_id/:protocol_id", HandleWithSession(s.getTestRunProtocol))
	api.GET("/tests", HandleWithSession(s.getTests))

	return nil
}

func (s *ApiService) socketHandler(c *gin.Context, w http.ResponseWriter, r *http.Request) {
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
