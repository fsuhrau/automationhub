package hub

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device"
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
}

func NewDeviceManager(logger *logrus.Logger, db *gorm.DB) *DeviceManager {
	return &DeviceManager{log: logger.WithFields(logrus.Fields{
		"prefix": "dm",
	}),
		db:             db,
		deviceHandlers: make(map[string]device.Handler),
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

/*
func (dm *DeviceManager) GetDevice(id uint) (device.Device, error) {
	var dev models.Device
	if err := dm.db.First(&dev, dev).Error; err != nil {
		return nil, err
	}
	for _, m := range dm.deviceHandlers {
		devices, _ := m.GetDevices()
		for _, i := range devices {
			if i.DeviceID() == dev.DeviceIdentifier {
				return i, nil
			}
		}
	}
	return nil, fmt.Errorf("device not found")
}
*/

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

/*
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

	// if len(properties.Architecture) > 0 && dev.() != properties.OS {
	// 	return false
	// }

	return true
}
*/
/*
func (dm *DeviceManager) LockDevice(session manager.Session, properties *device.Properties) (*manager.DeviceLock, error) {
	for _, mng := range dm.deviceHandlers {
		devices, _ := mng.GetDevices()
		for i := range devices {

			metsRequirements := evaluateDevice(devices[i], properties)

			if metsRequirements && !dm.isLocked(devices[i]) {
				dm.log.Debugf("LockDevice %v for session %s", devices[i], session.GetSessionID())
				lock := &manager.DeviceLock{
					Device: devices[i],
				}
				dm.LockedDevices[session.GetSessionID()] = lock
				session.SetDeviceLock(lock)
				return lock, nil
			}
		}
	}
	return nil, device.ManagerNotFoundError
}

func (dm *DeviceManager) UnlockDevice(session manager.Session) error {

	if d, ok := dm.LockedDevices[session.GetSessionID()]; ok {
		dm.log.Debugf("UnockDevice %v from session %s", d, session.GetSessionID())
		if d.Device.Connection() != nil  {
			dm.log.Debugf("Cose Connection for device %v from session %s", d, session.GetSessionID())
			d.Device.Connection().Close()
		}
		if err := d.Device.StopApp(session.GetAppParameter()); err != nil {
			logrus.Errorf("Stop App failed: %v", err)
		}

		dm.Stop(d.Device)

		delete(dm.LockedDevices, session.GetSessionID())
		return nil
	}
	return fmt.Errorf("no locked device found for session")
}
*/

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
					dm.log.Errorf("refresh devices failed for manager %s: %v", m.Name(), err)
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()
	return nil
}

func (dm *DeviceManager) updateDeviceState(dev device.Device) {
	deviceData := models.Device{
		DeviceIdentifier: dev.DeviceID(),
	}
	if err:= dm.db.FirstOrCreate(&deviceData, "device_identifier = ?", dev.DeviceID()).Error; err != nil {
		logrus.Errorf("%v", err)
		return
	}
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

/*
func (dm *DeviceManager) lookupDevice(connect *action.Connect, remoteAddress string) *manager.DeviceLock {

	if connect == nil {
		return nil
	}

	// by DEVICE_ID
	if len(connect.DeviceID) > 0 {
		for _, dev := range dm.LockedDevices {
			if dev.Device.DeviceID() == connect.DeviceID {
				return dev
			}
		}
	}

	// by SESSION_ID
	if len(connect.SessionID) > 0 {
		if dev, ok := dm.LockedDevices[connect.SessionID]; ok {
			return dev
		}
	}

	// by ip
	for _, dev := range dm.LockedDevices {
		if strings.Contains(remoteAddress, dev.Device.DeviceIP().String()) {
			return dev
		}
	}

	// not found ...
	return nil
}
*/

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
			ActionChannel:        make(chan action.Response, 1),
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

	dm.log.Debugf("Send Action: %s %v", reflect.TypeOf(act).Elem().Name(), act)

	buf, err := act.Serialize()
	if err != nil {
		return fmt.Errorf("Could not marshal Action: %v", err)
	}

	if dev.Connection() == nil {
	return fmt.Errorf("device not connected")
}

	dev.Log("send action: %v", act)
	return dev.Connection().Send(buf)
}

func (dm *DeviceManager) handleActions(d device.Device) {
	dm.log.Info("handleActions")

	defer func() {
		dm.log.Info("handleActions finished")
		d.Connection().Close()
		d.SetConnection(nil)
	}()

	for {
		select {
		case data := <-d.Connection().ResponseChannel:
			if data.Err == nil {
				resp := action.Response{}
				_ = proto.Unmarshal(data.Data, &resp)
				if d.ActionHandler() != nil {
					d.ActionHandler().OnActionResponse(d, &resp)
				} else {
					if resp.ActionID != "" {
						d.Connection().ActionChannel <- resp
					} else {
						d.Log("Received: %v", resp.Payload)
					}
				}
			}
		case status := <-d.Connection().ConnectionStateChannel:
			if status == device.Disconnected {
				dm.log.Info("handleActions disconnect")
				if d.ActionHandler() != nil {
					d.ActionHandler().OnActionResponse(d, nil)
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
