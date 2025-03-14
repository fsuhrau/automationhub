package web

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"net/http"
	"path/filepath"
	"sync"
)

var once sync.Once

type wf struct {
	cfg         config.Service
	filesystem  http.FileSystem
	indexBuffer []byte
	db          *gorm.DB
}

func New(db *gorm.DB, config config.Service) *wf {
	return &wf{
		cfg: config,
		db:  db,
	}
}

func (s *wf) IndexRoute(context *gin.Context) {
	once.Do(func() {
		file := "/data/index.html"
		f, _ := s.filesystem.Open(file)
		s.indexBuffer, _ = io.ReadAll(f)
	})
	context.Data(200, "text/html", s.indexBuffer)
}

func (s *wf) StaticRoute(context *gin.Context) {
	url := context.Request.RequestURI
	file := filepath.Join("data", url)
	context.FileFromFS(file, s.filesystem)
}

func (s *wf) RegisterRoutes(r *gin.Engine) error {
	s.filesystem = http.FS(Content)
	var (
		sessionFunc gin.HandlerFunc
		authFunc    gin.HandlerFunc
		authRouter  *gin.RouterGroup
		authRoutes  func(group *gin.Engine)
	)

	/*

		TODO check authentication
			if s.cfg.Auth.AuthenticationRequired() {
				if s.cfg.Auth.OAuth2 != nil {
					oauth2.Setup(s.cfg.Auth.OAuth2.RedirectUrl, s.cfg.Auth.OAuth2.AuthUrl, s.cfg.Auth.OAuth2.TokenUrl, s.cfg.Auth.OAuth2.UserUrl, s.cfg.Auth.OAuth2.Credentials, s.cfg.Auth.Github.Scopes, []byte(s.cfg.Auth.Github.Secret))
					sessionFunc = oauth2.Session("session")
					authFunc = oauth2.Auth()
				} else if s.cfg.Auth.Github != nil {
					github.Setup(s.cfg.Auth.Github.RedirectUrl, s.cfg.Auth.Github.Credentials, s.cfg.Auth.Github.Scopes, []byte(s.cfg.Auth.Github.Secret))
					sessionFunc = github.Session("session")
					authFunc = github.Auth()
				} else if s.cfg.Auth.Password != nil {
					password.Setup(s.db, []byte(s.cfg.Auth.Password.Secret))
					sessionFunc = password.Session("session")
					authFunc = password.Auth()
					authRoutes = password.Routes
				}
			}
	*/
	if sessionFunc != nil {
		r.Use(sessionFunc)
	}

	if authRoutes != nil {
		authRoutes(r)
	}

	authRouter = r.Group("/")
	if authFunc != nil {
		authRouter.Use(authFunc)
	}

	uploads := authRouter.Group("/upload")
	uploads.Static("/", apps.AppStoragePath)

	data := authRouter.Group("/data")
	data.Static("/", apps.TestDataPath)

	r.GET("/static/*filepath", s.StaticRoute)
	r.GET("/asset-manifest.json", s.StaticRoute)
	r.GET("/favicon.ico", s.StaticRoute)
	r.GET("/logo192.png", s.StaticRoute)
	r.GET("/logo512.png", s.StaticRoute)
	r.GET("/manifest.json", s.StaticRoute)
	r.NoRoute(s.IndexRoute)
	return nil
}
