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

func (s *wf) RegisterRoutes(r *gin.Engine, auth *gin.RouterGroup) error {
	s.filesystem = http.FS(Content)

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
