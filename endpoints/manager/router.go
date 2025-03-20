package manager

import (
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/config/protocol"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  protocol.SocketFrameSize,
	WriteBufferSize: protocol.SocketFrameSize,
}

type Service struct {
	logger         *logrus.Entry
	devicesManager manager.Devices
	cfg            config.Service
}

func New(logger *logrus.Logger, devicesManager manager.Devices, cfg config.Service) *Service {
	return &Service{
		logger:         logger.WithField("Service", "Devices"),
		cfg:            cfg,
		devicesManager: devicesManager,
	}
}

func (s *Service) socketHandler(c *gin.Context, w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Fail to get register request: %+v", err)
		return
	}

	s.registerDevices(msg, conn, c)
}

func (s *Service) RegisterRoutes(r *gin.Engine) error {
	deviceApi := r.Group("/manager")
	deviceApi.GET("/register", func(c *gin.Context) {
		s.socketHandler(c, c.Writer, c.Request)
	})

	return nil
}

func (s *Service) registerDevices(msg []byte, conn *websocket.Conn, c *gin.Context) {

	type Request struct {
		Type         string
		DeviceID     string
		IP           string
		Version      string
		OS           string
		UnityVersion string
		Name         string
		DeviceModel  string
		RAM          float32
		GPU          string
		SOC          string
		ProjectDir   string
	}

	clientIp := net.ParseIP(c.ClientIP())

	var req Request

	if err := json.Unmarshal(msg, &req); err != nil {
		return
	}

	register := device.RegisterData{
		DeviceOSVersion: req.Version,
		Name:            req.Name,
		DeviceOS:        req.OS,
		DeviceID:        req.DeviceID,
		TargetVersion:   req.UnityVersion,
		ManagerType:     req.Type,
		GPU:             req.GPU,
		RAM:             req.RAM,
		SOC:             req.SOC,
		DeviceModel:     req.DeviceModel,
		ProjectDir:      req.ProjectDir,
		DeviceIP:        clientIp,
		Conn:            conn,
	}

	_, _ = s.devicesManager.RegisterDevice(register)

}
