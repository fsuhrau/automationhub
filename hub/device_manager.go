package hub

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"io"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/golang/protobuf/proto"

	"github.com/sirupsen/logrus"
)

const (
	ReceiveBufferSize       = 20 * 1024
	DeviceConnectionTimeout = 5 * time.Second
	DefaultSocketTimeout    = 60 * time.Second
)

var (
	DeviceDisconnectedError = fmt.Errorf("device disconnected")
)

type DeviceManager struct {
	Managers      map[string]device.Handler
	LockedDevices map[string]*manager.DeviceLock
	stop          bool
	log           *logrus.Entry
}

func NewManager(logger *logrus.Logger) *DeviceManager {
	return &DeviceManager{log: logger.WithFields(logrus.Fields{
		"prefix": "dm",
	}),
		Managers:      make(map[string]device.Handler),
		LockedDevices: make(map[string]*manager.DeviceLock),
	}
}

func (dm *DeviceManager) AddHandler(manager device.Handler) {
	dm.Managers[manager.Name()] = manager
}

func (dm *DeviceManager) ListDevices() {
	for _, m := range dm.Managers {
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
	for _, m := range dm.Managers {
		d, err := m.GetDevices()
		if err != nil {
			dm.log.Errorf("failed: %v", err)
		}
		devices = append(devices, d...)
	}
	return devices, nil
}

func (dm *DeviceManager) Start(dev device.Device) error {
	for _, m := range dm.Managers {
		if m.HasDevice(dev) {
			return m.StartDevice(dev.DeviceID())
		}
	}
	return device.DeviceNotFoundError
}

func (dm *DeviceManager) Stop(dev device.Device) error {
	for _, m := range dm.Managers {
		if m.HasDevice(dev) {
			return m.StopDevice(dev.DeviceID())
		}
	}
	return device.DeviceNotFoundError
}

func (dm *DeviceManager) isLocked(dev device.Device) bool {
	for _, v := range dm.LockedDevices {
		if v.Device == dev {
			return true
		}
	}
	return false
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

	// if len(properties.Architecture) > 0 && dev.() != properties.OS {
	// 	return false
	// }

	return true
}

func (dm *DeviceManager) LockDevice(session manager.Session, properties *device.Properties) (*manager.DeviceLock, error) {
	for _, mng := range dm.Managers {
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
		if d.Connection != nil {
			dm.log.Debugf("Cose Connection for device %v from session %s", d, session.GetSessionID())
			d.Connection.Close()
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

func (dm *DeviceManager) Run(ctx context.Context) error {
	dm.log.Debug("Starting device manager")

	for _, v := range dm.Managers {
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
			for _, m := range dm.Managers {
				if err := m.RefreshDevices(); err != nil {
					dm.log.Errorf("refresh devices failed for manager %s: %v", m.Name(), err)
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()
	return nil
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

func (dm *DeviceManager) lookupDevice(session *action.Session, remoteAddress string) *manager.DeviceLock {

	if session == nil {
		return nil
	}

	// by DEVICE_ID
	if len(session.DeviceID) > 0 {
		for _, dev := range dm.LockedDevices {
			if dev.Device.DeviceID() == session.DeviceID {
				return dev
			}
		}
	}

	// by SESSION_ID
	if len(session.SessionID) > 0 {
		if dev, ok := dm.LockedDevices[session.SessionID]; ok {
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

func (dm *DeviceManager) handleConnection(c net.Conn) {
	if err := c.SetDeadline(time.Now().Add(DefaultSocketTimeout)); err != nil {
		dm.log.Errorf("SocketAccept SetDeadline: %v", err)
		return
	}

	remoteAddress := c.RemoteAddr().String()
	dm.log.Infof("SocketAccept established: %v", remoteAddress)
	var content string
	chunkBuffer := make([]byte, 4)
	_, err := c.Read(chunkBuffer)
	if err != nil {
		dm.log.Errorf("SocketAccept ReadError: %v", err)
		return
	}
	r := bytes.NewReader(chunkBuffer)
	var messageSize uint32
	binary.Read(r, binary.LittleEndian, &messageSize)
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
	session := resp.GetSession()

	content = string(buffer)
	dm.log.Debugf("SocketAccept content: %s", content)
	if lock := dm.lookupDevice(session, remoteAddress); lock != nil {
		dm.log.Infof("Received Handshake from %v", remoteAddress)
		dm.log.Debugf("Device with ID %s connected", lock.Device.DeviceID())
		lock.Connection = c
		lock.Device.SetConnectionState(device.Connected)
		lock.ResponseChannel = make(chan manager.ResponseData, 1)
		//lock.ConnectionStateChannel = make(chan bool, 1)
		go handleMessages(dm.log, lock)
	} else {
		dm.log.Errorf("no devide found for session: %s and device: %s", session.SessionID, session.DeviceID)
	}
}

func handleMessages(log *logrus.Entry, dev *manager.DeviceLock) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	for {
		if err := dev.Connection.SetDeadline(time.Now().Add(DefaultSocketTimeout)); err != nil {
			log.Errorf("SocketAccept SetDeadline: %v", err)
			return
		}
		var responseData manager.ResponseData
		chunkBuffer := make([]byte, 4)
		_, err := dev.Connection.Read(chunkBuffer)
		if err != nil {
			if io.EOF != err {
				log.Info("Device disconnected: %v", err)
			} else {
				log.Info("Device disconnected")
			}
			responseData.Err = DeviceDisconnectedError
			// dev.ConnectionStateChannel <- true

			if dev.Connection != nil {
				dev.Connection.Close()
			}
			if dev.ResponseChannel != nil {
				dev.ResponseChannel <- responseData
			}
			dev.Device.SetConnectionState(device.Disconnected)
			if dev.WaitingGroup != nil {
				dev.WaitingGroup.Done()
			}
			close(dev.ResponseChannel)
			dev.Connection = nil
			dev.WaitingGroup = nil
			return
		}

		r := bytes.NewReader(chunkBuffer)
		var messageSize uint32
		binary.Read(r, binary.LittleEndian, &messageSize)
		responseData.Data = make([]byte, 0, messageSize)
		chunkBuffer = make([]byte, ReceiveBufferSize)
		for uint32(len(responseData.Data)) < messageSize {
			n, err := dev.Connection.Read(chunkBuffer)
			if err != nil {
				responseData.Err = err
				log.Errorf("Chunk ReadError: %v", err)
				break
			}
			responseData.Data = append(responseData.Data, chunkBuffer[:n]...)
		}

		dev.ResponseChannel <- responseData
		if dev.WaitingGroup != nil {
			dev.WaitingGroup.Done()
		}
	}
}

func (dm *DeviceManager) StopObserver() {
	dm.stop = true
}

func (dm *DeviceManager) SendBytes(session manager.Session, content []byte) ([]byte, error) {
	start := time.Now()
	var size []byte
	size = make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(content)))
	defer func() {
		elapsed := time.Since(start)
		session.GetLogger().WithField("prefix", "dm").Debugf("send message took %s content: %s", elapsed.String(), content)
	}()

	if session.GetDeviceLock().Connection != nil {
		session.GetDeviceLock().WaitingGroup = &sync.WaitGroup{}
		session.GetDeviceLock().WaitingGroup.Add(1)
		session.GetDeviceLock().Connection.Write(size)
		session.GetDeviceLock().Connection.Write(content)
		session.GetDeviceLock().WaitingGroup.Wait()
		session.GetDeviceLock().WaitingGroup = nil
		data := <-session.GetDeviceLock().ResponseChannel
		if data.Err == DeviceDisconnectedError {
			session.HandleDisconnect()
		}
		return data.Data, data.Err
	}

	return []byte{}, fmt.Errorf("device not connected")
}

func (dm *DeviceManager) Send(session manager.Session, content string) ([]byte, error) {
	return dm.SendBytes(session, []byte(content))
}

func (dm *DeviceManager) SendAction(logger *logrus.Entry, session manager.Session, act action.Interface) error {
	logger.Debugf("Send Action: %s %v", reflect.TypeOf(act).Elem().Name(), act)
	buf, err := act.Serialize()
	if err != nil {
		return fmt.Errorf("Could not marshal Action: %v", err)
	}
	response, err := dm.SendBytes(session, buf)
	if err != nil {
		return err
	}
	logger.Debugf("Deserialize Action")
	return act.Deserialize(response)
}
