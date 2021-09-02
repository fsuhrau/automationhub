package api

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"sort"
	"time"
)

type ApiService struct {
	logger          *logrus.Entry
	db              *gorm.DB
	devicesManager  manager.Devices
	sessionsManager manager.Sessions
}

func New(logger *logrus.Logger, db *gorm.DB, dm manager.Devices, sm manager.Sessions) *ApiService {
	return &ApiService{
		logger:          logger.WithField("Service", "Api"),
		db:              db,
		devicesManager:  dm,
		sessionsManager: sm,
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
	api := r.Group("/api", func(context *gin.Context) {
	})
	if true {
		api.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000", "https://automationhub.com"},
			AllowMethods:     []string{"PUT", "POST", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
			ExposeHeaders:    []string{"Content-Length", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	} else {
		api.Use(cors.Default())
	}

	api.GET("/devices", func(c *gin.Context) {
		devices, _ := s.devicesManager.Devices()
		list := Selectables{}
		for i := range devices {
			connectionStatus := "Disconnected"
			if devices[i].Connection() != nil {
				connectionStatus = "Connected"
			}
			list = append(list, &Selectable{
				devices[i].DeviceID(),
				devices[i].DeviceName(),
				devices[i].DeviceOSName(),
				devices[i].DeviceOSVersion(),
				device.StateToString(devices[i].DeviceState()),
				connectionStatus,
			})
		}
		sort.Sort(ByName{list})
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, list)
	})

	api.GET("/tests", HandleWithSession(s.getTests))
	api.POST("/test", HandleWithSession(s.newTest))
	api.GET("/test/:test_id", HandleWithSession(s.getTest))
	api.POST("/test/:test_id/run", HandleWithSession(s.runTest))
	api.GET("/test/:test_id/runs", HandleWithSession(s.getTestRuns))
	api.GET("/test/:test_id/runs/:run_id", HandleWithSession(s.getTestRun))

	api.POST("/execute", HandleWithSession(s.startTest))
	api.GET("/status", HandleWithSession(s.getStatus))
	return nil
}
