package web

import (
	"github.com/fsuhrau/automationhub/authentication/github"
	"github.com/fsuhrau/automationhub/authentication/oauth2"
	"github.com/fsuhrau/automationhub/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type wf struct {
	cfg config.Service
}

func New(config config.Service) *wf {
	return &wf{
		cfg: config,
	}
}

func (s *wf) RegisterRoutes(r *gin.Engine) error {
	fs := http.FS(Content)

	if s.cfg.Auth.AuthenticationRequired() {
		if s.cfg.Auth.OAuth2 != nil {
			oauth2.Setup(s.cfg.Auth.OAuth2.RedirectUrl, s.cfg.Auth.OAuth2.AuthUrl, s.cfg.Auth.OAuth2.TokenUrl, s.cfg.Auth.OAuth2.UserUrl, s.cfg.Auth.OAuth2.Credentials, s.cfg.Auth.Github.Scopes, []byte(s.cfg.Auth.Github.Secret))
			r.Use(oauth2.Session("session"))
		} else if s.cfg.Auth.Github != nil {
			github.Setup(s.cfg.Auth.Github.RedirectUrl, s.cfg.Auth.Github.Credentials, s.cfg.Auth.Github.Scopes, []byte(s.cfg.Auth.Github.Secret))
			r.Use(github.Session("session"))
		}
	}

	authorized := r.Group("/web")
	uploads := r.Group("/upload")

	if s.cfg.Auth.AuthenticationRequired() {
		if s.cfg.Auth.OAuth2 != nil {
			authorized.Use(oauth2.Auth())
			uploads.Use(oauth2.Auth())
		} else if s.cfg.Auth.Github != nil {
			authorized.Use(github.Auth())
			uploads.Use(github.Auth())
		}
	}

	uploads.Static("/", "./upload")

	authorized.GET("/*filepath", func(context *gin.Context) {
		url := context.Request.RequestURI
		file := strings.Replace(url, "/web/", "/data/", -1)
		_, err := fs.Open(file)
		if err != nil {
			file = "/data/"
		}
		context.FileFromFS(file, fs)
	})

	r.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/web")
	})
	return nil
}
