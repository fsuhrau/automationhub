package hub

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/endpoints"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"time"

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

	router.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowMethods:     []string{"POST, OPTIONS, GET, PUT", "DELETE"},
		AllowOrigins:     []string{"http://10.35.111.51:3000", "http://localhost:3000", "http://localhost:8002", "https://automationhub.com"},
		AllowHeaders:     []string{"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Access-Control-Allow-Origin"},
		MaxAge:           12 * time.Hour,
	}))

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
