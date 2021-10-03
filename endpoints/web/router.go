package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type wf struct {
}

func New() *wf {
	return &wf{}
}

func (s *wf) RegisterRoutes(r *gin.Engine) error {
	fs := http.FS(Content)
	if false {
		r.GET("/", func(context *gin.Context) {
			context.FileFromFS("/", fs)
		})
		r.GET("/static/*filepath", func(context *gin.Context) {
			context.FileFromFS(context.Request.RequestURI, fs)
		})
		r.GET("/asset-manifest.json", func(context *gin.Context) {
			context.FileFromFS("/asset-manifest.json", fs)
		})
		r.GET("/favicon.ico", func(context *gin.Context) {
			context.FileFromFS("/favicon.ico", fs)
		})
		r.GET("/logo192.png", func(context *gin.Context) {
			context.FileFromFS("/logo192.png", fs)
		})
		r.GET("/logo512.png", func(context *gin.Context) {
			context.FileFromFS("/logo512.png", fs)
		})
		r.GET("/manifest.json", func(context *gin.Context) {
			context.FileFromFS("/manifest.json", fs)
		})
		r.GET("/robots.txt", func(context *gin.Context) {
			context.FileFromFS("/robots.txt", fs)
		})

	} else {
		r.GET("/*filepath", func(context *gin.Context) {
			context.FileFromFS(context.Request.RequestURI, fs)
		})
	}

	return nil
}
