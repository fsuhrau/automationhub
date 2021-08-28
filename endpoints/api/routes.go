package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"sort"
)

type ApiService struct {
	logger          *logrus.Entry
	devicesManager  manager.Devices
	sessionsManager manager.Sessions
}

func New(logger *logrus.Logger, dm manager.Devices, sm manager.Sessions) *ApiService {
	return &ApiService{
		logger: logger.WithField("Service", "Api"),
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
			list = append(list, &Selectable{devices[i].DeviceID(), fmt.Sprintf("%s(%s) %s", devices[i].DeviceOSName(), devices[i].DeviceOSVersion(), devices[i].DeviceName())})
		}
		sort.Sort(ByName{list})
		c.JSON(http.StatusOK, list)
	})

	api.GET("tests", HandleWithSession(s.getTests))
	api.POST("execute", HandleWithSession(s.startTest))
	api.GET("status", HandleWithSession(s.getStatus))
	return nil
}