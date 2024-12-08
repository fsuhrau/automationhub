package hub

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/endpoints"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/modules/hooks"
	"github.com/fsuhrau/automationhub/modules/hooks/notifier"
	"github.com/fsuhrau/automationhub/modules/hooks/slack"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type Service struct {
	server        *http.Server
	deviceManager manager.Devices
	logger        *logrus.Logger
	// hostIP        net.IP
	endpoints []endpoints.ServiceEndpoint
	router    *gin.Engine
	cfg       config.Service
	hooks     []hooks.Hook
	sd        storage.Device
	db        *gorm.DB
}

func NewService(logger *logrus.Logger, ip net.IP, devices manager.Devices, cfg config.Service, sd storage.Device, db *gorm.DB) *Service {
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
		AllowOrigins:     []string{"http://localhost:7109", "http://localhost:3000", "http://localhost:8002", "http://10.35.111.12"},
		AllowHeaders:     []string{"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With", "X-Auth-Token"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Access-Control-Allow-Origin"},
		MaxAge:           12 * time.Hour,
	}))

	return &Service{logger: logger, deviceManager: devices, router: router, cfg: cfg, sd: sd, db: db}
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

func (s *Service) RegisterHooks(hooksCfgs []config.Hook) {
	for _, hook := range hooksCfgs {
		if hook.Provider == "slack" {
			s.hooks = append(s.hooks, slack.NewHook(hook))
		}
	}

	notifier.RegisterEventTestRunFinishedListener(s.hooks)
}
