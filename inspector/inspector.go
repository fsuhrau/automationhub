package inspector

import (
	"github.com/fsuhrau/automationhub/inspector/handler/dashboard"
	"github.com/fsuhrau/automationhub/inspector/handler/gui"
	"github.com/fsuhrau/automationhub/inspector/renderer"
	"github.com/gin-gonic/gin"
)

func Init(router *gin.Engine) {
	render := renderer.New()
	render.Debug = gin.IsDebugging()
	render.Layout = "layouts/index"
	render.TemplatesDir = "inspector/handler/"
	router.HTMLRender = render.Create()
	router.Static("/images", "./public/images")
	router.Static("/css", "./public/css")
	router.Static("/js", "./public/js")
	router.GET("/", dashboard.Index)
	router.GET("/inspector", gui.Index)
	router.POST("/inspector/start", gui.Start)
}