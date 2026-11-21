package hub

import (
	"net/http"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/endpoints"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/middleware"
	"github.com/fsuhrau/automationhub/modules/hooks"
	"github.com/fsuhrau/automationhub/modules/hooks/notifier"
	"github.com/fsuhrau/automationhub/modules/hooks/slack"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// Service is the http server of a master or a node: the router, the
// endpoints mounted on it and the device manager it runs.
type Service struct {
	deviceManager       manager.Devices
	logger              *logrus.Logger
	endpoints           []endpoints.ServiceEndpoint
	publicRouter        *gin.Engine
	authenticatedRouter *gin.RouterGroup
	cfg                 config.Service
	hooks               []hooks.Hook
}

// Deps are what NewService is built on.
type Deps struct {
	Logger  *logrus.Logger
	Devices manager.Devices
	Config  config.Service
	// Auth guards the authenticated routes; the zero value leaves them open.
	Auth Auth
}

// NewService builds the router with the CORS, session and guard middleware of
// the deps and applies the log level of the config to the logger.
func NewService(deps Deps) *Service {
	applyLogLevel(deps.Logger, deps.Config.LogLevel)
	deps.Logger.Formatter = new(prefixed.TextFormatter)
	router := newRouter(deps.Logger)
	router.Use(cors.New(middleware.CORSConfig(deps.Config.Cors)))
	return &Service{
		logger:              deps.Logger,
		deviceManager:       deps.Devices,
		publicRouter:        router,
		authenticatedRouter: deps.Auth.mount(router),
		cfg:                 deps.Config,
	}
}

// applyLogLevel sets the level named in the config, info when it is empty or invalid.
func applyLogLevel(logger *logrus.Logger, name string) {
	level, err := logrus.ParseLevel(name)
	if err != nil {
		logrus.Infof("Parse Log Level: %s", err)
		logrus.Info("using default info")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
}

func newRouter(logger *logrus.Logger) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger(logger.WithFields(logrus.Fields{"prefix": "service"})))
	return r
}

// Handler is the router of the service, for serving it without Run.
func (s *Service) Handler() http.Handler { return s.publicRouter }

// AddEndpoint registers the routes of the endpoint.
func (s *Service) AddEndpoint(endpoint endpoints.ServiceEndpoint) error {
	if err := endpoint.RegisterRoutes(s.publicRouter, s.authenticatedRouter); err != nil {
		return err
	}
	s.endpoints = append(s.endpoints, endpoint)
	return nil
}

// RegisterHooks creates the hooks of the config (only slack is known) and
// lets them hear about finished test runs.
func (s *Service) RegisterHooks(hooksCfgs []config.Hook) {
	for _, hook := range hooksCfgs {
		if hook.Provider == "slack" {
			s.hooks = append(s.hooks, slack.NewHook(hook))
		}
	}

	notifier.RegisterEventTestRunFinishedListener(s.hooks, s.cfg.HostIP)
}

// openSession answers the session check of a hub without authentication.
func openSession(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
