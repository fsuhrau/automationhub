package hub

import (
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/authentication/github"
	"github.com/fsuhrau/automationhub/authentication/oauth2"
	"github.com/fsuhrau/automationhub/authentication/oauthflow"
	"github.com/fsuhrau/automationhub/authentication/password"
	"github.com/fsuhrau/automationhub/authentication/token"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/endpoints"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/middleware"
	"github.com/fsuhrau/automationhub/modules/hooks"
	"github.com/fsuhrau/automationhub/modules/hooks/notifier"
	"github.com/fsuhrau/automationhub/modules/hooks/slack"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type Service struct {
	server        *http.Server
	deviceManager manager.Devices
	logger        *logrus.Logger
	// hostIP        net.IP
	endpoints           []endpoints.ServiceEndpoint
	publicRouter        *gin.Engine
	authenticatedRouter *gin.RouterGroup
	cfg                 config.Service
	hooks               []hooks.Hook
	sd                  storage.Device
	db                  *gorm.DB
}

func NewService(logger *logrus.Logger, ip net.IP, devices manager.Devices, cfg config.Service, sd storage.Device, db *gorm.DB) *Service {
	applyLogLevel(logger, cfg.LogLevel)
	logger.Formatter = new(prefixed.TextFormatter)
	router := newRouter(logger)

	router.Use(cors.New(middleware.CORSConfig(cfg.Cors)))

	var (
		sessionFunc    gin.HandlerFunc
		sessionHandler gin.HandlerFunc
		authRouter     *gin.RouterGroup
		authRoutes     func(group *gin.Engine)
	)

	if cfg.Auth.AuthenticationRequired() {
		sessionFunc, sessionHandler, authRoutes = setupAuth(db, cfg.Auth)
	} else {
		router.GET("/auth/session", func(context *gin.Context) {
			context.JSON(200, gin.H{"status": "ok"})
		})
	}

	if sessionFunc != nil {
		router.Use(sessionFunc)
	}

	if authRoutes != nil {
		authRoutes(router)
	}

	authRouter = router.Group("/")
	if sessionHandler != nil {
		authRouter.Use(token.SessionHandler(db, sessionHandler))
	}

	return &Service{logger: logger, deviceManager: devices, publicRouter: router, authenticatedRouter: authRouter, cfg: cfg, sd: sd, db: db}
}

// setupAuth configures the first configured authentication provider
// (OAuth2, GitHub, password) and returns its session middleware, session
// handler and routes. A provider that can't be set up ends the process.
func setupAuth(db *gorm.DB, auth config.Auth) (sessionFunc, sessionHandler gin.HandlerFunc, authRoutes func(*gin.Engine)) {
	provider, err := authProvider(db, auth)
	if err != nil {
		logrus.Fatalf("Failed to set up authentication: %v", err)
	}
	if provider == nil {
		return nil, nil, nil
	}
	return provider.Session("session"), provider.SessionHandler(), provider.Routes
}

// authProvider builds the provider of the first configured method, nil for none.
func authProvider(db *gorm.DB, auth config.Auth) (authentication.Provider, error) {
	switch {
	case auth.OAuth2 != nil:
		o := auth.OAuth2
		endpoints := oauth2.Endpoints{AuthURL: o.AuthUrl, TokenURL: o.TokenUrl, UserURL: o.UserUrl}
		return oauth2.New(db, endpoints, oauthflow.Settings{RedirectURL: o.RedirectUrl, CredFile: o.Credentials, Scopes: o.Scopes, Secret: []byte(o.Secret)})
	case auth.Github != nil:
		g := auth.Github
		return github.New(db, oauthflow.Settings{RedirectURL: g.RedirectUrl, CredFile: g.Credentials, Scopes: g.Scopes, Secret: []byte(g.Secret)})
	case auth.Password != nil:
		return password.New(db, []byte(auth.Password.Secret)), nil
	}
	return nil, nil
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

func (s *Service) AddEndpoint(endpoint endpoints.ServiceEndpoint) error {
	if err := endpoint.RegisterRoutes(s.publicRouter, s.authenticatedRouter); err != nil {
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

	notifier.RegisterEventTestRunFinishedListener(s.hooks, s.cfg.HostIP)
}
