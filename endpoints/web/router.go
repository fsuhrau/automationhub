package web

import (
	"fmt"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"net/http"
	"path/filepath"
)

type wf struct {
	cfg         config.Service
	filesystem  http.FileSystem
	indexBuffer []byte
	db          *gorm.DB
}

func New(db *gorm.DB, config config.Service) *wf {
	return &wf{
		cfg:        config,
		db:         db,
		filesystem: http.FS(Content),
	}
}

func (s *wf) loadIndex() error {
	f, err := s.filesystem.Open("/data/index.html")
	if err != nil {
		return fmt.Errorf("open index.html: %w", err)
	}
	defer func() { _ = f.Close() }()
	s.indexBuffer, err = io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read index.html: %w", err)
	}
	return nil
}

func (s *wf) IndexRoute(context *gin.Context) {
	context.Data(200, "text/html", s.indexBuffer)
}

func (s *wf) StaticRoute(context *gin.Context) {
	url := context.Request.RequestURI
	file := filepath.Join("data", url)
	context.FileFromFS(file, s.filesystem)
}

func (s *wf) RegisterRoutes(r *gin.Engine, auth *gin.RouterGroup) error {
	if err := s.loadIndex(); err != nil {
		return err
	}

	uploads := auth.Group("/upload")
	uploads.Static("/", apps.AppStoragePath)
	data := auth.Group("/data")
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
