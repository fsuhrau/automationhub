package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/config/protocol"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"reflect"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"

	"github.com/fsuhrau/automationhub/hub/action"
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
}

func NewDeviceManager(logger *logrus.Logger, masterUrl, nodeIdentifier string) *DeviceManager {
	return &DeviceManager{log: logger.WithFields(logrus.Fields{
		"prefix": "dm",
	}),
		masterUrl:      masterUrl,
		nodeIdentifier: nodeIdentifier,
		deviceHandlers: make(map[string]device.Handler),
		deviceCache:    make(map[string]*models.Device),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  protocol.SocketFrameSize,
			WriteBufferSize: protocol.SocketFrameSize,
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
			return m.StartDevice(dev.DeviceID())
		}
	}
	return device.DeviceNotFoundError
}

func (dm *DeviceManager) Stop(dev device.Device) error {
	for _, m := range dm.deviceHandlers {
		if m.HasDevice(dev) {
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

	for _, v := range dm.deviceHandlers {
		v.Init(dm.masterUrl, dm.nodeIdentifier)
		v.Start()
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				dm.log.Infof("Stop DeviceManager")
				return
			default:
			}
			// dm.log.Debugf("refreshing device lists...")
			for _, m := range dm.deviceHandlers {
				if err := m.RefreshDevices(false); err != nil {
					// dm.log.Errorf("refresh devices failed for manager %s: %v", m.Name(), err)
				}
			}
		}
	}()
	return nil
}

/*
func (dm *DeviceManager) getDevice(deviceID string) *models.Device {
	if m, ok := dm.deviceCache[deviceID]; ok {
		return m
	}

	deviceData := models.Device{
		DeviceIdentifier: deviceID,
		Status:           device.StateUnknown,
	}

	if err := dm.db.FirstOrCreate(&deviceData, "device_identifier = ?", deviceID).Error; err != nil {
		logrus.Errorf("fail to create new device: %v", err)
	}
	dm.deviceCache[deviceID] = &deviceData
	return dm.deviceCache[deviceID]
}
*/

func (dm *DeviceManager) RegisterRoutes(r *gin.Engine) error {
	deviceApi := r.Group("/device")
	deviceApi.GET("/connect", func(c *gin.Context) {
		conn, err := dm.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("Failed to set websocket upgrade: %+v", err)
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

	defer func() {
		dm.log.Info("handleActions finished")
		if d.Connection() != nil {
			if d.Connection().ResponseChannel != nil {
				close(d.Connection().ResponseChannel)
			}
			d.Connection().Close()
		}
		d.SetConnection(nil)
	}()

	for {
		select {
		case data := <-d.Connection().ResponseChannel:
			if data.Err == nil {
				resp := action.Response{}
				if err := json.Unmarshal(data.Data, &resp); err != nil {
					logrus.Error(err)
				}
				if resp.ActionType == action.ActionType_Log {
					logError(d, resp.Payload.LogData)
					continue
				}
				if handler := d.ActionHandlers(); handler != nil {
					for i := range handler {
						handler[i].OnActionResponse(d, &resp)
					}
					continue
				}
				if resp.ActionType == action.ActionType_Performance {
					continue
				}
				if resp.ActionID != "" {
					d.Connection().ActionChannel <- resp
					continue
				}
				if resp.ActionType == action.ActionType_ExecuteMethodStart {
					continue
				}
				if resp.ActionType == action.ActionType_ExecuteMethodFinished {
					continue
				}
				if resp.ActionType == action.ActionType_ExecuteTest {
					continue
				}
				if resp.ActionType == action.ActionType_ExecutionResult {
					continue
				}

				d.Log("unhandled", "Received: %v", resp.Payload)
				return
			} else {
				return
			}
		case <-ctx.Done():
			dm.log.Info("handleActions cancel")
			if handler := d.ActionHandlers(); handler != nil {
				for i := range handler {
					handler[i].OnActionResponse(d, nil)
				}
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
