package api

import (
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"sort"
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
		if s == nil {
			return
		}
		session := s.(*Session)
		// session.LastAccess = time.Now()
		c.Header("Access-Control-Allow-Origin", "*")
		f(session, c)
	}
}

func (s *ApiService) RegisterRoutes(r *gin.Engine) error {
	api := r.Group("/api", func(context *gin.Context) {
	})

	api.GET("/devices", func(c *gin.Context) {
		devices, _ := s.devicesManager.Devices()
		list := Selectables{}
		for i := range devices {
			list = append(list, &Selectable{
				devices[i].DeviceID(),
				devices[i].DeviceName(),
				devices[i].DeviceOSName(),
				devices[i].DeviceOSVersion(),
				"",
			})
		}
		sort.Sort(ByName{list})
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, list)
	})

	api.GET("/tests", HandleWithSession(s.getTests))
	api.POST("/test", HandleWithSession(s.newTest))
	api.GET("/test/:test_id", HandleWithSession(s.getTest))
	api.GET("/test/:test_id/runs", HandleWithSession(s.getTestRuns))
	api.GET("/test/:test_id/runs/:run_id", HandleWithSession(s.getTestRun))

	api.POST("/execute", HandleWithSession(s.startTest))
	api.GET("/status", HandleWithSession(s.getStatus))
	return nil
}
