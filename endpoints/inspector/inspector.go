package inspector

import (
	"github.com/fsuhrau/automationhub/config/protocol"
	"github.com/fsuhrau/automationhub/endpoints/inspector/handler/dashboard"
	"github.com/fsuhrau/automationhub/endpoints/inspector/handler/devices"
	"github.com/fsuhrau/automationhub/endpoints/inspector/handler/gui"
	"github.com/fsuhrau/automationhub/endpoints/inspector/handler/sessions"
	"github.com/fsuhrau/automationhub/endpoints/inspector/renderer"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Inspector struct {
	logger   *logrus.Entry
	devices  manager.Devices
	sessions manager.Sessions
}

func New(logger *logrus.Logger, dm manager.Devices, sm manager.Sessions) *Inspector {
	return &Inspector{
		logger:   logger.WithField("service", "inspector"),
		devices:  dm,
		sessions: sm,
	}
}

func (s *Inspector) RegisterRoutes(r *gin.Engine) error {
	render := renderer.New()
	render.Debug = gin.IsDebugging()
	render.Layout = "layouts/index"
	render.TemplatesDir = "endpoints/inspector/handler/"
	r.HTMLRender = render.Create()
	r.Static("/images", "./public/images")
	r.Static("/css", "./public/css")
	r.Static("/js", "./public/js")
	r.GET("/", dashboard.Index(s.devices, s.sessions))
	r.GET("/inspector", gui.Index(s.devices))
	r.POST("/inspector/upload", gui.UploadFile)
	r.GET("/inspector/devices", devices.Index(s.devices))
	r.GET("/inspector/sessions", sessions.Index(s.sessions))
	r.GET("/inspector/session/:sessionID/show", sessions.Show(s.sessions))
	r.GET("/ws", func(c *gin.Context) {
		wshandler(c.Writer, c.Request)
	})
	//router.GET("/tests", gui.Tests)
	return nil
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  protocol.SocketFrameSize,
	WriteBufferSize: protocol.SocketFrameSize,
}

func wshandler(w http.ResponseWriter, r *http.Request) {
	wsupgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("Socket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	for {
		t, msg, err := conn.ReadMessage()
		logrus.Info(msg)
		if err != nil {
			break
		}
		conn.WriteMessage(t, msg)
	}
}
