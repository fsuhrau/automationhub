package hub

import (
	"github.com/fsuhrau/automationhub/endpoints"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type Service struct {
	server         *http.Server
	deviceManager  manager.Devices
	sessionManager manager.Sessions
	logger         *logrus.Logger
	hostIP         net.IP
	endpoints      []endpoints.ServiceEndpoint
	router         *gin.Engine
	// sessions       map[string]*Session
}

func NewService(logger *logrus.Logger, devices manager.Devices, sessions manager.Sessions) *Service {
	level, err := logrus.ParseLevel(viper.GetString("log"))
	if err != nil {
		logrus.Infof("Parse Log Level: %s", err)
		logrus.Info("using default info")
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(level)
	}
	logger.Formatter = new(prefixed.TextFormatter)
	router := newRouter(logger)
	return &Service{logger: logger, sessionManager: sessions, deviceManager: devices, router: router}
}

func newRouter(logger *logrus.Logger) *gin.Engine {
	r := gin.New()
	r.Use(Recovery())
	r.Use(Logger(logger.WithFields(logrus.Fields{"prefix": "service"})))
	return r
}

func (s *Service) AddEndpoint(endpoint endpoints.ServiceEndpoint) error {
	if err := endpoint.RegisterRoutes(s.router); err != nil {
		return err
	}

	s.endpoints = append(s.endpoints, endpoint)
	return nil
}
