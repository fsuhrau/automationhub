package inspector

import (
	"github.com/fsuhrau/automationhub/inspector/handler/dashboard"
	"github.com/fsuhrau/automationhub/inspector/handler/devices"
	"github.com/fsuhrau/automationhub/inspector/handler/gui"
	"github.com/fsuhrau/automationhub/inspector/handler/manager"
	"github.com/fsuhrau/automationhub/inspector/handler/sessions"
	"github.com/fsuhrau/automationhub/inspector/renderer"
	"github.com/gin-gonic/gin"
)

func Init(router *gin.Engine, dm manager.DeviceManager, sm manager.SessionManager) {
	render := renderer.New()
	render.Debug = gin.IsDebugging()
	render.Layout = "layouts/index"
	render.TemplatesDir = "inspector/handler/"
	router.HTMLRender = render.Create()
	router.Static("/images", "./public/images")
	router.Static("/css", "./public/css")
	router.Static("/js", "./public/js")
	router.GET("/", dashboard.Index(dm, sm))
	router.GET("/inspector", gui.Index(dm))
	router.POST("/inspector/upload", gui.UploadFile)
	router.GET("/inspector/devices", devices.Index(dm))
	router.GET("/inspector/sessions", sessions.Index(sm))
	router.GET("/inspector/session/:sessionID/show", sessions.Show(sm))
	//router.GET("/tests", gui.Tests)
}
