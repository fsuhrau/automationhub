package inspector

import (
	"github.com/fsuhrau/automationhub/inspector/handler/dashboard"
	"github.com/fsuhrau/automationhub/inspector/handler/devices"
	"github.com/fsuhrau/automationhub/inspector/handler/gui"
	"github.com/fsuhrau/automationhub/inspector/handler/manager"
	"github.com/fsuhrau/automationhub/inspector/handler/sessions"
	"github.com/fsuhrau/automationhub/inspector/renderer"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
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
	router.GET("/ws", func (c *gin.Context) {
		wshandler(c.Writer, c.Request)
	})
	//router.GET("/tests", gui.Tests)
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}
func wshandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("Socket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		conn.WriteMessage(t, msg)
	}
}