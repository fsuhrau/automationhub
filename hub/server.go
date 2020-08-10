package hub

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device/androiddevice"
	"github.com/fsuhrau/automationhub/device/iosdevice"
	"github.com/fsuhrau/automationhub/device/iossim"
	"github.com/fsuhrau/automationhub/device/macos"
	"github.com/fsuhrau/automationhub/inspector"
	"github.com/fsuhrau/automationhub/remlog"
	"github.com/gin-gonic/gin"
	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)


type Selectable struct {
	ID string
	Name string
}
type Selectables []*Selectable
func (s Selectables) Len() int      { return len(s) }
func (s Selectables) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByName struct{ Selectables }
func (s ByName) Less(i, j int) bool { return s.Selectables[i].Name < s.Selectables[j].Name }


type Server struct {
	server        *http.Server
	deviceManager *DeviceManager
	sessions      map[string]*Session
	logger        *logrus.Logger
	hostIP net.IP
}

func NewServer() *Server {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.Formatter = new(prefixed.TextFormatter)
	return &Server{logger: logger, sessions: make(map[string]*Session), deviceManager: NewManager(logger)}
}

func GetOutboundIP() net.IP {
	// return net.IPv4(0,0,0,0)

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

func (s *Server) Sessions() map[string]*Session {
	return s.sessions
}

func (s *Server) cleanupSessions() {
	for sessionID, session := range s.sessions {
		if session.LastAccess.Add(time.Duration(10 * time.Second)).Before(time.Now()) {
			s.logger.Debugf("session %s expired", sessionID)

			if session.Recorder != nil {
				if err := session.Recorder.Stop(); err != nil {
					logrus.Errorf("stop recording session failed: %v", err)
				}
			}

			s.deviceManager.UnlockDevice(session)
			delete(s.sessions, session.SessionID)
		}
	}
}

func (s *Server) Run() error {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.hostIP = GetOutboundIP()

	//gocron.Every(30).Second().Do(s.cleanupSessions)

	// f, _ := os.OpenFile("debug.log", os.O_WRONLY|os.O_CREATE, 0755)
	// defer f.Close()
	// s.logger.SetOutput(f)

	go ZeroConfServer(ctx, "", s.hostIP.String())

	remoteLoggingService := remlog.NewService(s.logger)
	if err := remoteLoggingService.Run(func(msg string) {
		data := strings.Split(msg, "|")
		if len(data) > 1 {
			if session, ok := s.sessions[data[0]]; ok {
				session.Storage.RemoteDeviceLog(data[1])
				session.logger.Info(data[1])
			} else {
				s.logger.Info(msg)
			}
		}
	}); err != nil {
		return err
	}

	// start device observer thread
	s.deviceManager.AddManager(iossim.NewManager(s.hostIP))
	s.deviceManager.AddManager(macos.NewManager(s.hostIP))
	s.deviceManager.AddManager(iosdevice.NewManager())
	s.deviceManager.AddManager(androiddevice.NewManager())
	if err := s.deviceManager.Run(ctx); err != nil {
		return err
	}

	r := gin.New()
	r.Use(Logger(s.logger.WithFields(logrus.Fields{"prefix": "rest"})), gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	inspector.Init(r, s.deviceManager)

	r.GET("/devices", func(c *gin.Context) {
		devices, _:= s.deviceManager.Devices()
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
