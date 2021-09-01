package hub

import (
	"github.com/fsuhrau/automationhub/config"
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
	cfg  config.Service
	// sessions       map[string]*Session
}

func NewService(logger *logrus.Logger, ip net.IP, devices manager.Devices, sessions manager.Sessions, cfg config.Service) *Service {
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
	return &Service{logger: logger, hostIP: ip, sessionManager: sessions, deviceManager: devices, router: router, cfg: cfg}
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
