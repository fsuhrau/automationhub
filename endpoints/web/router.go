package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type wf struct {
}

func New() *wf {
	return &wf{}
}

func (s *wf) RegisterRoutes(r *gin.Engine) error {
	fs := http.FS(Content)
	//authorized := r.Group("/web", gin.BasicAuth(gin.Accounts{
	//	"admin": "password",
	//}))

	authorized := r.Group("/web")

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
