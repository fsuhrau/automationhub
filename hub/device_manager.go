package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/config/protocol"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"reflect"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"

	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/sirupsen/logrus"
)

const (
	DeviceConnectionTimeout = 5 * time.Minute
)

type DeviceManager struct {
	db             *gorm.DB
	deviceHandlers map[string]device.Handler
	stop           bool
	log            *logrus.Entry
	deviceCache    map[string]*models.Device
	upgrader       websocket.Upgrader
	masterUrl      string
	nodeIdentifier string
	authToken      *string

	refreshSignal chan struct{}

	// pushes device changes to the master; nil on the master itself
	notifier *node.DevicesNotifier
}

func NewDeviceManager(logger *logrus.Logger, service config.Service) *DeviceManager {
	var authToken *string
	if service.Auth.Token != nil {
		authToken = &service.Auth.Token.AuthToken
	}
	return &DeviceManager{log: logger.WithFields(logrus.Fields{
		"prefix": "dm",
	}),
		masterUrl:      service.MasterURL,
		nodeIdentifier: service.Identifier,
		authToken:      authToken,
		deviceHandlers: make(map[string]device.Handler),
		deviceCache:    make(map[string]*models.Device),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  protocol.SocketFrameSize,
			WriteBufferSize: protocol.SocketFrameSize,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				// mobile has no origin find a better check
				if origin == "" {
					return true
				}

				for i := range service.Cors {
					if service.Cors[i] == origin {
						return true
					}
				}

				return false
			},
		},
	}
}

func (dm *DeviceManager) Managers() []string {
	var manager []string
	for k, _ := range dm.deviceHandlers {
		manager = append(manager, k)
	}
	return manager
}

func (dm *DeviceManager) AddHandler(manager device.Handler) {
	dm.deviceHandlers[manager.Name()] = manager
}

func (dm *DeviceManager) ListDevices() {
	for _, m := range dm.deviceHandlers {
		dm.log.Infof("Manager: %s", m.Name())
		devices, err := m.GetDevices()
		if err != nil {
			dm.log.Errorf("failed: %v", err)
		}
		for _, d := range devices {
			dm.log.Infof("Device %s (%s) %d %s", d.DeviceName(), d.DeviceID(), d.DeviceState(), d.DeviceOSName())
		}
	}
}

func (dm *DeviceManager) Devices() (map[string][]device.Device, error) {
	var devices map[string][]device.Device
	devices = make(map[string][]device.Device)

	for h, m := range dm.deviceHandlers {
		d, err := m.GetDevices()
		if err != nil {
			dm.log.Errorf("failed: %v", err)
		}
		devices[h] = d
	}
	return devices, nil
}

func (dm *DeviceManager) Start(dev device.Device) error {
	for _, m := range dm.deviceHandlers {
		if m.HasDevice(dev) {
			defer dm.notifier.Notify()
			return m.StartDevice(dev.DeviceID())
		}
	}
	return device.DeviceNotFoundError
}

func (dm *DeviceManager) Stop(dev device.Device) error {
	for _, m := range dm.deviceHandlers {
		if m.HasDevice(dev) {
			defer dm.notifier.Notify()
			return m.StopDevice(dev.DeviceID())
		}
	}
	return device.DeviceNotFoundError
}

func evaluateDevice(dev device.Device, properties *device.Properties) bool {
	if properties == nil {
		return true
	}

	if len(properties.DeviceID) > 0 && dev.DeviceID() != properties.DeviceID {
		return false
	}

	if len(properties.Name) > 0 && dev.DeviceName() != properties.Name {
		return false
	}

	if len(properties.OS) > 0 && dev.DeviceOSName() != properties.OS {
		return false
	}

	return true
}

func (dm *DeviceManager) FindDevice(params *app.Parameter) device.Device {
	for _, mng := range dm.deviceHandlers {
		devices, _ := mng.GetDevices()
		for i := range devices {
			if devices[i].DeviceOSName() == params.Platform {
				return devices[i]
			}
		}
	}
	return nil
}

func (dm *DeviceManager) RegisterDevice(data device.RegisterData) (device.Device, error) {
	for _, mng := range dm.deviceHandlers {
		if mng.Name() == data.ManagerType {
			return mng.RegisterDevice(data)
		}
	}
	return nil, fmt.Errorf("device for type not found")
}

func (dm *DeviceManager) Run(ctx context.Context, runSocketListener bool) error {
	dm.log.Debug("Starting device manager")

	dm.notifier = node.NewDevicesNotifier(dm.masterUrl, dm.nodeIdentifier, dm.authToken, dm.deviceSnapshot)
	for _, v := range dm.deviceHandlers {
		if n, ok := v.(device.ChangeNotifying); ok {
			n.SetChangeNotifier(dm.notifier.Notify)
		}
		v.Init(dm.masterUrl, dm.nodeIdentifier, dm.authToken)
		v.Start()
	}
	// the initial list, so the master is up to date without asking
	dm.notifier.Notify()

	refreshDevicesTicker := time.NewTicker(5 * time.Second)
	//defer refreshDevicesTicker.Stop()

	dm.refreshSignal = make(chan struct{}, 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				dm.log.Infof("Stop DeviceManager")
				refreshDevicesTicker.Stop()
				return
			case <-refreshDevicesTicker.C:
				dm.refreshDeviceList()
			case <-dm.refreshSignal:
				dm.refreshDeviceList()
			}
		}
	}()

	return nil
}

func (dm *DeviceManager) refreshDeviceList() {
	for _, m := range dm.deviceHandlers {
		if err := m.RefreshDevices(false); err != nil {
			// dm.log.Errorf("refresh devices failed for manager %s: %v", m.Name(), err)
		}
	}
	// handlers without OS events (simulators, editors, ...) and state changes
	// made by sessions only show up here; the notifier sends nothing when the
	// list is unchanged
	dm.notifier.Notify()
}

func (dm *DeviceManager) deviceSnapshot() []*node.DeviceResponse {
	devices, _ := dm.Devices()
	return node.ToDevicesResponse(devices)
}

func (dm *DeviceManager) TriggerRefresh() {
	select {
	case dm.refreshSignal <- struct{}{}:
	default: // Avoid blocking if the channel is full
	}
}

func (dm *DeviceManager) RegisterRoutes(r *gin.Engine, auth *gin.RouterGroup) error {
	deviceApi := auth.Group("/device")
	deviceApi.GET("/connect", func(c *gin.Context) {
		conn, err := dm.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Printf("Failed to set websocket upgrade: %+v\n", err)
			return
		}

		dm.handleConnection(conn)
	})

	return nil
}

func (dm *DeviceManager) handleConnection(conn *websocket.Conn) {
	// defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(DeviceConnectionTimeout)); err != nil {
		dm.log.Errorf("WebSocketAccept SetDeadline: %v", err)
	}

	remoteAddress := conn.RemoteAddr().String()
	dm.log.Infof("WebSocketAccept established: %v", remoteAddress)

	_, buffer, err := conn.ReadMessage()
	if err != nil {
		dm.log.Errorf("WebSocketAccept: %+v", err)
		return
	}

	resp := &action.Response{}
	if err := json.Unmarshal(buffer, resp); err != nil {
		dm.log.Errorf("SocketAccept ReadError: %v", err)
		return
	}

	dev, _ := dm.GetDevice(resp.Payload.Connect.DeviceID)
	if dev != nil {
		dm.log.Infof("Received Handshake from %v", remoteAddress)
		dm.log.Debugf("Device with ID %s connected", dev.DeviceID())
		connection := &device.Connection{
			ConnectionParameter: resp.Payload.Connect,
			Logger:              dm.log,
			Connection:          conn,
			ResponseChannel:     make(chan device.ResponseData, 100),
			ActionChannel:       make(chan action.Response, 1),
		}
		ctx := dev.NewContext()
		dev.SetConnection(connection)
		go connection.HandleMessages(ctx)
		go dm.handleActions(dev, ctx)
	}
}

func (dm *DeviceManager) StopObserver() {
	dm.stop = true
}

func (dm *DeviceManager) SendAction(dev device.Device, act action.Interface) error {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		dm.log.WithField("prefix", "dm").Debugf("send message took %s", elapsed.String())
	}()

	buf, err := act.Serialize()
	if err != nil {
		return fmt.Errorf("Could not marshal Action: %v", err)
	}

	if !dev.IsAppConnected() {
		return fmt.Errorf("device not connected")
	}

	dm.log.Debugf("Send Action: %s %v", reflect.TypeOf(act).Elem().Name(), act)
	dev.Log("action", "Send Action: %s", reflect.TypeOf(act).Elem().Name())
	return dev.Send(buf)
}

func getTypeOfLog(logType action.LogType) string {
	switch logType {
	case action.LogType_DeviceLog:
		return "app"
	case action.LogType_StepLog:
		return "step"
	case action.LogType_StatusLog:
		return "status"
	case action.LogType_CheckpointLog:
		return "checkpoint"
	case action.LogType_PerformanceLog:
		return "performance"
	}
	return ""
}

func getLevelOfLog(loglevel action.LogLevel) string {
	switch loglevel {
	case action.LogLevel_Debug:
		return "debug"
	case action.LogLevel_Info:
		return "info"
	case action.LogLevel_Warning:
		return "warning"
	case action.LogLevel_Error:
		return "error"
	case action.LogLevel_Exception:
		return "exception"
	}
	return ""
}

func logError(dev device.Device, log *action.LogData) bool {
	if log.Level == action.LogLevel_Exception {
		dev.Exception(getTypeOfLog(log.Type), log.Message)
		return true
	} else if log.Level == action.LogLevel_Error {
		dev.Error(getTypeOfLog(log.Type), log.Message)
	} else {
		dev.Log(getTypeOfLog(log.Type), log.Message)
	}
	return false
}

func (dm *DeviceManager) handleActions(d device.Device, ctx context.Context) {
	dm.log.Info("handleActions")

	con := d.Connection()
	if con == nil {
		return
	}

	defer func() {
		dm.log.Info("handleActions finished")
		con.Close()
		// do not reset a newer connection that replaced this one
		if d.Connection() == con {
			d.SetConnection(nil)
		}
	}()

	for {
		select {
		case data, ok := <-con.ResponseChannel:
			if !ok {
				dm.log.Info("channel closed")
				return
			}
			if data.Err != nil {
				return
			}

			resp := action.Response{}
			if err := json.Unmarshal(data.Data, &resp); err != nil {
				dm.log.Errorf("invalid action response from %s: %v", d.DeviceID(), err)
				continue
			}
			if resp.ActionType == action.ActionType_Log {
				if resp.Payload.LogData != nil {
					logError(d, resp.Payload.LogData)
				}
				continue
			}
			if handlers := d.ActionHandlers(); len(handlers) > 0 {
				for i := range handlers {
					handlers[i].OnActionResponse(d, &resp)
				}
				continue
			}
			if resp.ActionType == action.ActionType_Performance {
				continue
			}
			if resp.ActionID != "" {
				// nobody is required to consume this channel; never block the read loop
				select {
				case con.ActionChannel <- resp:
				default:
					dm.log.Debugf("dropping unhandled action response %s (%v)", resp.ActionID, resp.ActionType)
				}
				continue
			}

			d.Log("unhandled", "Received: %v", resp.Payload)
		case <-ctx.Done():
			dm.log.Info("handleActions cancel")
			// executors registered on the device stay registered across app
			// reconnects; only a cancelled device context aborts them
			for _, handler := range d.ActionHandlers() {
				handler.OnActionResponse(d, nil)
			}
			return
		}
	}
}

func (dm *DeviceManager) GetDevice(id string) (device.Device, string) {
	for manager, mng := range dm.deviceHandlers {
		devices, _ := mng.GetDevices()
		for i := range devices {
			if devices[i].DeviceID() == id {
				return devices[i], manager
			}
		}
	}

	return nil, ""
}
