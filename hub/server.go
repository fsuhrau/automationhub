package hub

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device/androiddevice"
	"github.com/fsuhrau/automationhub/device/iosdevice"
	"github.com/fsuhrau/automationhub/device/iossim"
	"github.com/fsuhrau/automationhub/device/macos"
	"github.com/fsuhrau/automationhub/inspector"
	"github.com/fsuhrau/automationhub/remlog"
	"github.com/gin-gonic/gin"
	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type Selectable struct {
	ID   string
	Name string
}
type Selectables []*Selectable

func (s Selectables) Len() int      { return len(s) }
func (s Selectables) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByName struct{ Selectables }

func (s ByName) Less(i, j int) bool { return s.Selectables[i].Name < s.Selectables[j].Name }

type Server struct {
	server         *http.Server
	deviceManager  *DeviceManager
	sessionManager *SessionManager
	logger         *logrus.Logger
	hostIP         net.IP
	// sessions       map[string]*Session
}

func NewServer() *Server {
	logger := logrus.New()
	level, err := logrus.ParseLevel(viper.GetString("log"))
	if err != nil {
		logrus.Infof("Parse Log Level: %s", err)
		logrus.Info("using default info")
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(level)
	}
	logger.Formatter = new(prefixed.TextFormatter)
	deviceManager := NewManager(logger)
	sessionManager := NewSessionManager(logger, deviceManager)
	return &Server{logger: logger /*sessions: make(map[string]*Session)*/, sessionManager: sessionManager, deviceManager: deviceManager}
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func ZeroConfServer(ctx context.Context, name string, address string) {
	logrus.Infof("Starting MDNS server")
	ifaces, err := net.Interfaces()
	if err != nil {
		logrus.Errorf("getting interfaces failed: %v", err)
	}

	allowedInterfaces := []net.Interface{}
	for _, interf := range ifaces {
		addresses, _ := interf.Addrs()
		for _, addr := range addresses {
			if strings.Contains(addr.String(), address) {
				allowedInterfaces = append(allowedInterfaces, interf)
				break
			}
		}
	}

	serviceName := "_automationhub._tcp"
	if name != "" {
		serviceName = fmt.Sprintf("_automationhub_%s._tcp", name)
	}

	server, err := zeroconf.Register("AutomationHub", serviceName, "local.", 3939, []string{"txtv=0", "lo=1", "la=2"}, allowedInterfaces)
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	// Clean exit.
	select {
	case <-ctx.Done():
	}

	logrus.Infof("MDNS server shutdown.")
}

func (s *Server) Run() error {

	showRemlog := viper.GetBool("display_remlog")

	var serviceConfig config.Service

	if err := viper.Unmarshal(&serviceConfig); err != nil {
		return err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if len(serviceConfig.IP) > 0 {
		s.hostIP = net.ParseIP(serviceConfig.IP)
	}

	if s.hostIP == nil {
		s.hostIP = GetOutboundIP()
	}

	go ZeroConfServer(ctx, "", s.hostIP.String())

	remoteLoggingService := remlog.NewService(s.logger)
	if err := remoteLoggingService.Run(func(msg string) {
		data := strings.Split(msg, "|")
		if len(data) > 1 {
			session, _ := s.sessionManager.GetSession(data[0])
			if session != nil {
				session.Storage.RemoteDeviceLog(data[1])
				if showRemlog {
					session.logger.Debug(data[1])
				}
			} else {
				if showRemlog {
					s.logger.Debug(msg)
				}
			}
		}
	}); err != nil {
		return err
	}

	// start device observer thread
	s.deviceManager.AddManager(iossim.NewManager(&serviceConfig, s.hostIP))
	s.deviceManager.AddManager(macos.NewManager(&serviceConfig, s.hostIP))
	s.deviceManager.AddManager(iosdevice.NewManager(&serviceConfig))
	s.deviceManager.AddManager(androiddevice.NewManager(&serviceConfig))
	if err := s.deviceManager.Run(ctx); err != nil {
		return err
	}

	s.sessionManager.Run(ctx)

	r := gin.New()
	r.Use(Logger(s.logger.WithFields(logrus.Fields{"prefix": "rest"})), gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	inspector.Init(r, s.deviceManager)

	r.GET("/devices", func(c *gin.Context) {
		devices, _ := s.deviceManager.Devices()
		list := Selectables{}
		for i := range devices {
			list = append(list, &Selectable{devices[i].DeviceID(), fmt.Sprintf("%s(%s) %s", devices[i].DeviceOSName(), devices[i].DeviceOSVersion(), devices[i].DeviceName())})
		}
		sort.Sort(ByName{list})
		c.JSON(http.StatusOK, list)
	})

	r.POST("/wd/hub/session", s.InitNewTestSession)
	authGroup := r.Group("/wd/hub/session/:sessionID")
	authGroup.Use(SessionMiddleware(s))
	authGroup.DELETE("", HandleWithSession(s.StopTestingSession))
	authGroup.GET("screenshot", HandleWithSession(s.TakeScreenshot))
	authGroup.GET("screen", HandleWithSession(s.GetScreen))
	authGroup.GET("graph", HandleWithSession(s.GetGraph))
	authGroup.POST("timeouts", HandleWithSession(s.SetTimeouts))
	authGroup.POST("element", HandleWithSession(s.GetElement))
	authGroup.POST("elements", HandleWithSession(s.GetElements))

	authGroup.GET("element/:elementID/attribute/:attribute", HandleWithSession(s.ElementGetAttribute))
	authGroup.GET("element/:elementID/displayed", HandleWithSession(s.ElementIsDisplayed))
	authGroup.GET("element/:elementID/text", HandleWithSession(s.ElementGetText))
	authGroup.GET("element/:elementID/css", HandleWithSession(s.ElementGetCSS))
	authGroup.GET("element/:elementID/name", HandleWithSession(s.ElementGetName))

	authGroup.POST("element/:elementID/click", HandleWithSession(s.ElementClick))
	authGroup.POST("element/:elementID/value", HandleWithSession(s.ElementSetValue))
	authGroup.POST("touch/longclick", HandleWithSession(s.LongClickElement))
	authGroup.POST("touch/position", HandleWithSession(s.TouchPosition))
	authGroup.POST("touch/down", HandleWithSession(s.TouchDown))
	authGroup.POST("touch/move", HandleWithSession(s.TouchMove))
	authGroup.POST("touch/up", HandleWithSession(s.TouchUp))
	authGroup.POST("moveto", HandleWithSession(s.MoveTo))
	authGroup.POST("buttondown", HandleWithSession(s.ButtonDown))
	authGroup.POST("buttonup", HandleWithSession(s.ButtonUp))
	authGroup.POST("url", HandleWithSession(s.RestartApp))
	authGroup.POST("back", HandleWithSession(s.NavigateBack))

	// runOn := fmt.Sprintf("%s:8002", ip.String())
	runOn := ":8002"
	err := r.Run(runOn)
	logrus.Infof("Stopping Server")
	if err != nil {
		return err
	}

	s.deviceManager.StopObserver()
	return nil
}
