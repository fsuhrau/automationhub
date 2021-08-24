package selenium

import (
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
)

type SeleniumService struct {
	devicesManager  manager.Devices
	sessionsManager manager.Sessions
	logger          *logrus.Entry
	hostIP         net.IP
}

func (s *SeleniumService) RegisterRoutes(r *gin.Engine, dm manager.Devices, sm manager.Sessions) error {
	s.devicesManager = dm
	s.sessionsManager = sm

	r.POST("/wd/hub/session", s.InitNewTestSession)
	authGroup := r.Group("/wd/hub/session/:sessionID")
	authGroup.Use(SessionMiddleware(s.sessionsManager))
	authGroup.DELETE("", HandleWithSession(s.StopTestingSession))
	authGroup.GET("ping", HandleWithSession(func(s *Session, c *gin.Context) {
		c.String(http.StatusOK, "pong")
	}))
	authGroup.GET("screenshot", HandleWithSession(s.TakeScreenshot))
	authGroup.GET("screen", HandleWithSession(s.GetScreen))
	authGroup.GET("graph", HandleWithSession(s.GetGraph))
	authGroup.POST("timeouts", HandleWithSession(s.SetTimeouts))
	authGroup.POST("element", HandleWithSession(s.GetElement))
	authGroup.POST("elements", HandleWithSession(s.GetElements))

	authGroup.GET("element/:elementID/attribute/:attribute", HandleWithSession(s.ElementGetAttribute))
	authGroup.GET("element/:elementID/displayed", HandleWithSession(s.ElementIsDisplayed))
	authGroup.GET("element/:elementID/text", HandleWithSession(s.ElementGetText))
	authGroup.GET("element/:elementID/css", HandleWithSession(s.ElementGetCSS))
	authGroup.GET("element/:elementID/name", HandleWithSession(s.ElementGetName))

	authGroup.POST("element/:elementID/click", HandleWithSession(s.ElementClick))
	authGroup.POST("element/:elementID/value", HandleWithSession(s.ElementSetValue))
	authGroup.POST("touch/longclick", HandleWithSession(s.LongClickElement))
	authGroup.POST("touch/position", HandleWithSession(s.TouchPosition))
	authGroup.POST("touch/down", HandleWithSession(s.TouchDown))
	authGroup.POST("touch/move", HandleWithSession(s.TouchMove))
	authGroup.POST("touch/up", HandleWithSession(s.TouchUp))
	authGroup.POST("moveto", HandleWithSession(s.MoveTo))
	authGroup.POST("buttondown", HandleWithSession(s.ButtonDown))
	authGroup.POST("buttonup", HandleWithSession(s.ButtonUp))
	authGroup.POST("url", HandleWithSession(s.RestartApp))
	authGroup.POST("back", HandleWithSession(s.NavigateBack))
	return nil
}