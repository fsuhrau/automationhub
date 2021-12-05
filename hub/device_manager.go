package hub

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
	"net"
	"reflect"
	"time"

	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/golang/protobuf/proto"

	"github.com/sirupsen/logrus"
)

const (
	DeviceConnectionTimeout = 30 * time.Minute
)

type DeviceManager struct {
	db             *gorm.DB
	deviceHandlers map[string]device.Handler
	stop           bool
	log            *logrus.Entry
	deviceCache    map[string]*models.Device
}

func NewDeviceManager(logger *logrus.Logger, db *gorm.DB) *DeviceManager {
	return &DeviceManager{log: logger.WithFields(logrus.Fields{
		"prefix": "dm",
	}),
		db:             db,
		deviceHandlers: make(map[string]device.Handler),
		deviceCache:    make(map[string]*models.Device),
	}
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

func (dm *DeviceManager) Devices() ([]device.Device, error) {
	var devices []device.Device
	for _, m := range dm.deviceHandlers {
		d, err := m.GetDevices()
		if err != nil {
			dm.log.Errorf("failed: %v", err)
		}
		devices = append(devices, d...)
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

func (dm *DeviceManager) Run(ctx context.Context) error {
	dm.log.Debug("Starting device manager")

	for _, v := range dm.deviceHandlers {
		v.Init()
		v.Start()
	}

	if err := dm.SocketListener(); err != nil {
		return err
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
				if err := m.RefreshDevices(dm.updateDeviceState); err != nil {
					// dm.log.Errorf("refresh devices failed for manager %s: %v", m.Name(), err)
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()
	return nil
}

func (dm *DeviceManager) getDevice(deviceID string) *models.Device {
	if m, ok := dm.deviceCache[deviceID]; ok {
		return m
	}

	deviceData := models.Device{
		DeviceIdentifier: deviceID,
		Status: device.StateUnknown,
	}

	if err := dm.db.FirstOrCreate(&deviceData, "device_identifier = ?", deviceID).Error; err != nil {
		logrus.Errorf("fail to create new device: %v", err)
	}
	dm.deviceCache[deviceID] = &deviceData
	return dm.deviceCache[deviceID]
}

func (dm *DeviceManager) updateDeviceState(dev device.Device) {
	deviceData := dm.getDevice(dev.DeviceID())

	needsUpdate := false
	if deviceData.Name != dev.DeviceName() {
		deviceData.Name = dev.DeviceName()
		needsUpdate = true
	}
	if deviceData.OS != dev.DeviceOSName() {
		deviceData.OS = dev.DeviceOSName()
		needsUpdate = true
	}
	if deviceData.OSVersion != dev.DeviceOSVersion() {
		deviceData.OSVersion = dev.DeviceOSVersion()
		needsUpdate = true
	}

	if deviceData.HardwareModel != dev.DeviceModel() {
		deviceData.HardwareModel = dev.DeviceModel()
		needsUpdate = true
	}

	statusUpdate := false
	if deviceData.Status != dev.DeviceState() {
		deviceData.Status = dev.DeviceState()
		statusUpdate = true
	}

	if needsUpdate || statusUpdate {
		dm.db.Updates(deviceData)
	}

	if statusUpdate {
		log := models.DeviceLog{
			DeviceID: deviceData.ID,
			Status:   dev.DeviceState(),
			Payload:  "",
		}
		dm.db.Create(&log)
		events.DeviceStatusChanged.Trigger(events.DeviceStatusChangedPayload{
			DeviceID:    deviceData.ID,
			DeviceState: uint(dev.DeviceState()),
		})
	}
}

func (dm *DeviceManager) SocketListener() error {
	// listen on all interfaces
	l, err := net.Listen("tcp", ":3939")
	if err != nil {
		return err
	}

	go func() {
		//defer l.Close()
		for {
			c, err := l.Accept()
			if err != nil {
				dm.log.Errorf("SocketAccept error: %v", err)
				return
			}

			dm.handleConnection(c)
		}
	}()
	return nil
}

func (dm *DeviceManager) handleConnection(c net.Conn) {
	if err := c.SetDeadline(time.Now().Add(DeviceConnectionTimeout)); err != nil {
		dm.log.Errorf("SocketAccept SetDeadline: %v", err)
		return
	}

	remoteAddress := c.RemoteAddr().String()
	dm.log.Infof("SocketAccept established: %v", remoteAddress)

	chunkBuffer := make([]byte, 4)
	_, err := c.Read(chunkBuffer)
	if err != nil {
		dm.log.Errorf("SocketAccept ReadError: %v", err)
		return
	}
	messageSize := device.GetMessageSize(chunkBuffer)
	dm.log.Infof("waiting for %d bytes", messageSize)

	buffer := make([]byte, 0, messageSize)
	for uint32(len(buffer)) < messageSize {
		n, err := c.Read(chunkBuffer)
		if err != nil {
			dm.log.Errorf("SocketAccept ReadError: %v", err)
			return
		}
		buffer = append(buffer, chunkBuffer[:n]...)
	}

	resp := &action.Response{}
	if err := proto.Unmarshal(buffer, resp); err != nil {
		dm.log.Errorf("SocketAccept ReadError: %v", err)
		return
	}
	connectRequest := resp.GetConnect()

	dev := dm.GetDevice(connectRequest.GetDeviceID())
	if dev != nil {
		dm.log.Infof("Received Handshake from %v", remoteAddress)
		dm.log.Debugf("Device with ID %s connected", dev.DeviceID())
		connection := &device.Connection{
			ConnectionParameter:    connectRequest,
			Logger:                 dm.log,
			Connection:             c,
			ResponseChannel:        make(chan device.ResponseData, 100),
			ActionChannel:          make(chan action.Response, 1),
			ConnectionStateChannel: make(chan device.ConnectionState, 1),
		}
		dev.SetConnection(connection)
		go connection.HandleMessages()
		go dm.handleActions(dev)
	}
	/*
		dm.log.Debugf("SocketAccept content: %s", content)
		if lock := dm.lookupDevice(session, remoteAddress); lock != nil {
			dm.log.Infof("Received Handshake from %v", remoteAddress)
			dm.log.Debugf("Device with ID %s connected", lock.Device.DeviceID())
		} else {
			dm.log.Errorf("no devide found for session: %s and device: %s", session.SessionID, session.DeviceID)
		}
	*/
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

	if dev.Connection() == nil {
		return fmt.Errorf("device not connected")
	}

	dm.log.Debugf("Send Action: %s %v", reflect.TypeOf(act).Elem().Name(), act)
	dev.Log("action", "Send Action: %s", reflect.TypeOf(act).Elem().Name())
	return dev.Connection().Send(buf)
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

func (dm *DeviceManager) handleActions(d device.Device) {
	dm.log.Info("handleActions")

	defer func() {
		dm.log.Info("handleActions finished")
		if d.Connection() != nil {
			d.Connection().Close()
		}
		d.SetConnection(nil)
	}()

	for {
		select {
		case data := <-d.Connection().ResponseChannel:
			if data.Err == nil {
				resp := action.Response{}
				_ = proto.Unmarshal(data.Data, &resp)
				actionHandler := d.ActionHandlers()
				if actionHandler != nil {
					if resp.ActionType == action.ActionType_Log && resp.GetLog() != nil {
						if resp.GetLog().Level >= action.LogLevel_Error {
							d.Error(getTypeOfLog(resp.GetLog().Type), resp.GetLog().Message)
						} else {
							d.Log(getTypeOfLog(resp.GetLog().Type), resp.GetLog().Message)
						}
					}
					for i := range actionHandler {
						actionHandler[i].OnActionResponse(d, &resp)
					}
				} else {
					if resp.ActionID != "" {
						d.Connection().ActionChannel <- resp
					} else {
						if resp.ActionType == action.ActionType_Log && resp.GetLog() != nil {
							if resp.GetLog().Level >= action.LogLevel_Error {
								d.Error(getTypeOfLog(resp.GetLog().Type), resp.GetLog().Message)
							} else {
								d.Log(getTypeOfLog(resp.GetLog().Type), resp.GetLog().Message)
							}
						} else {
							d.Log("unhandled", "Received: %v", resp.Payload)
						}
					}
				}
			}
		case status := <-d.Connection().ConnectionStateChannel:
			if status == device.Disconnected {
				dm.log.Info("handleActions disconnect")
				actionHandler := d.ActionHandlers()
				if actionHandler != nil {
					for i := range actionHandler {
						actionHandler[i].OnActionResponse(d, nil)
					}
				}
				return
			}
		}
	}
}

func (dm *DeviceManager) GetDevice(id string) device.Device {
	for _, mng := range dm.deviceHandlers {
		devices, _ := mng.GetDevices()
		for i := range devices {
			if devices[i].DeviceID() == id {
				return devices[i]
			}
		}
	}
	return nil
}
