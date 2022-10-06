package web

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/gin-gonic/gin"
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
}

func New(config config.Service) *wf {
	return &wf{
		cfg: config,
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
	/*
		if s.cfg.Auth.AuthenticationRequired() {
			if s.cfg.Auth.OAuth2 != nil {
				oauth2.Setup(s.cfg.Auth.OAuth2.RedirectUrl, s.cfg.Auth.OAuth2.AuthUrl, s.cfg.Auth.OAuth2.TokenUrl, s.cfg.Auth.OAuth2.UserUrl, s.cfg.Auth.OAuth2.Credentials, s.cfg.Auth.Github.Scopes, []byte(s.cfg.Auth.Github.Secret))
				r.Use(oauth2.Session("session"))
			} else if s.cfg.Auth.Github != nil {
				github.Setup(s.cfg.Auth.Github.RedirectUrl, s.cfg.Auth.Github.Credentials, s.cfg.Auth.Github.Scopes, []byte(s.cfg.Auth.Github.Secret))
				r.Use(github.Session("session"))
			}
		}
	*/
	uploads := r.Group("/upload")
	{
		/*
			authorized := r.Group("/")
			if s.cfg.Auth.AuthenticationRequired() {
				if s.cfg.Auth.OAuth2 != nil {
					authorized.Use(oauth2.Auth())
					uploads.Use(oauth2.Auth())
				} else if s.cfg.Auth.Github != nil {
					authorized.Use(github.Auth())
					uploads.Use(github.Auth())
				}
			}
		*/
		uploads.Static("/", apps.AppStoragePath)
	}
	data := r.Group("/data")
	{
		data.Static("/", apps.TestDataPath)
	}

	r.GET("/static/*filepath", s.StaticRoute)
	r.GET("/asset-manifest.json", s.StaticRoute)
	r.GET("/favicon.ico", s.StaticRoute)
	r.GET("/logo192.png", s.StaticRoute)
	r.GET("/logo512.png", s.StaticRoute)
	r.GET("/manifest.json", s.StaticRoute)
	r.NoRoute(s.IndexRoute)
	return nil
}
