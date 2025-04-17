package hub

import (
	"github.com/fsuhrau/automationhub/authentication/github"
	"github.com/fsuhrau/automationhub/authentication/oauth2"
	"github.com/fsuhrau/automationhub/authentication/password"
	"github.com/fsuhrau/automationhub/authentication/token"
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
	endpoints           []endpoints.ServiceEndpoint
	publicRouter        *gin.Engine
	authenticatedRouter *gin.RouterGroup
	cfg                 config.Service
	hooks               []hooks.Hook
	sd                  storage.Device
	db                  *gorm.DB
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
		AllowOrigins:     cfg.Cors,
		AllowHeaders:     []string{"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With", "X-Auth-Token"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Access-Control-Allow-Origin"},
		MaxAge:           12 * time.Hour,
	}))

	var (
		sessionFunc    gin.HandlerFunc
		sessionHandler gin.HandlerFunc
		authRouter     *gin.RouterGroup
		authRoutes     func(group *gin.Engine)
	)

	if cfg.Auth.AuthenticationRequired() {
		if cfg.Auth.OAuth2 != nil {
			oauth2.Setup(db, cfg.Auth.OAuth2.RedirectUrl, cfg.Auth.OAuth2.AuthUrl, cfg.Auth.OAuth2.TokenUrl, cfg.Auth.OAuth2.UserUrl, cfg.Auth.OAuth2.Credentials, cfg.Auth.Github.Scopes, []byte(cfg.Auth.Github.Secret))
			sessionFunc = oauth2.Session("session")
			sessionHandler = oauth2.SessionHandler()
			authRoutes = oauth2.Routes
		} else if cfg.Auth.Github != nil {
			github.Setup(db, cfg.Auth.Github.RedirectUrl, cfg.Auth.Github.Credentials, cfg.Auth.Github.Scopes, []byte(cfg.Auth.Github.Secret))
			sessionFunc = github.Session("session")
			sessionHandler = github.SessionHandler()
			authRoutes = github.Routes
		} else if cfg.Auth.Password != nil {
			password.Setup(db, []byte(cfg.Auth.Password.Secret))
			sessionFunc = password.Session("session")
			sessionHandler = password.SessionHandler()
			authRoutes = password.Routes
		}
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

func newRouter(logger *logrus.Logger) *gin.Engine {
	r := gin.New()
	r.Use(Recovery())
	r.Use(Logger(logger.WithFields(logrus.Fields{"prefix": "service"})))
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

	notifier.RegisterEventTestRunFinishedListener(s.hooks)
}
