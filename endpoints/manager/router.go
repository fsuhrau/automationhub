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

type Service struct {
	logger         *logrus.Entry
	devicesManager manager.Devices
	cfg            config.Service
	upgrader       websocket.Upgrader
}

func New(logger *logrus.Logger, devicesManager manager.Devices, cfg config.Service) *Service {
	return &Service{
		logger:         logger.WithField("Service", "Devices"),
		cfg:            cfg,
		devicesManager: devicesManager,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  protocol.SocketFrameSize,
			WriteBufferSize: protocol.SocketFrameSize,
		},
	}
}

func (s *Service) socketHandler(c *gin.Context, w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v\n", err)
		return
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Printf("Fail to get register request: %+v\n", err)
		return
	}

	s.registerDevices(msg, conn, c)
}

func (s *Service) RegisterRoutes(r *gin.Engine, auth *gin.RouterGroup) error {
	deviceApi := auth.Group("/manager")
	deviceApi.GET("/register", func(c *gin.Context) {
		s.socketHandler(c, c.Writer, c.Request)
	})

	return nil
}

type registerRequest struct {
	Type          string
	DeviceID      string
	IP            string
	Version       string
	OS            string
	TargetVersion string
	Name          string
	DeviceModel   string
	RAM           float32
	GPU           string
	SOC           string
	ProjectDir    string
}

func (r registerRequest) registerData(ip net.IP, conn *websocket.Conn) device.RegisterData {
	return device.RegisterData{
		DeviceOSVersion: r.Version,
		Name:            r.Name,
		DeviceOS:        r.OS,
		DeviceID:        r.DeviceID,
		TargetVersion:   r.TargetVersion,
		ManagerType:     r.Type,
		GPU:             r.GPU,
		RAM:             r.RAM,
		SOC:             r.SOC,
		DeviceModel:     r.DeviceModel,
		ProjectDir:      r.ProjectDir,
		DeviceIP:        ip,
		Conn:            conn,
	}
}

func (s *Service) registerDevices(msg []byte, conn *websocket.Conn, c *gin.Context) {
	var req registerRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		return
	}

	register := req.registerData(net.ParseIP(c.ClientIP()), conn)
	_, _ = s.devicesManager.RegisterDevice(register)
}
