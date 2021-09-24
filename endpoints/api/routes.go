package api

import (
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/sse"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net"
)

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
	api.DELETE("/app/:app_id", HandleWithSession(s.deleteApp))
	api.GET("/app/:app_id/functions", HandleWithSession(s.getAppFunctions))
	api.POST("/app/upload", HandleWithSession(s.uploadApp))
	api.GET("/apps", HandleWithSession(s.getApps))

	api.GET("/device/:device_id", HandleWithSession(s.getDeviceStatus))
	api.POST("/device/:device_id/tests", HandleWithSession(s.deviceRunTests))
	api.GET("/devices", HandleWithSession(s.getDevices))

	api.POST("/test", HandleWithSession(s.newTest))
	api.GET("/test/:test_id", HandleWithSession(s.getTest))
	api.POST("/test/:test_id/run", s.runTest)
	api.GET("/test/:test_id/runs", HandleWithSession(s.getTestRuns))
	api.GET("/test/:test_id/runs/last", HandleWithSession(s.getLastTestRun))
	api.GET("/test/:test_id/run/:run_id", HandleWithSession(s.getTestRun))
	api.GET("/tests", HandleWithSession(s.getTests))

	api.POST("/execute", HandleWithSession(s.startTest))
	api.GET("/status", HandleWithSession(s.getStatus))
	return nil
}
