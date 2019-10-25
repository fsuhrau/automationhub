package hub

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/devices"
	"github.com/sirupsen/logrus"
)

const (
	ReceiveBufferSize       = 20 * 1024
	DeviceConnectionTimeout = 5 * time.Second
)

type DeviceProperties struct {
	Name         string
	Type         string
	OS           string
	Architecture string
	App          string
	AppId        string
}

type DeviceLock struct {
	Device          devices.Device
	Connection      net.Conn
	AppName         string
	WaitingGroup    *sync.WaitGroup
	ResponseChannel chan []byte
}

type DeviceManager struct {
	Managers      map[string]devices.Manager
	LockedDevices map[string]*DeviceLock
	stop          bool
	log           *logrus.Entry
}

func NewManager(logger *logrus.Logger) *DeviceManager {

	return &DeviceManager{log: logger.WithFields(logrus.Fields{
		"prefix": "dm",
	}),
		Managers:      make(map[string]devices.Manager),
		LockedDevices: make(map[string]*DeviceLock),
	}
}

func (dm *DeviceManager) AddManager(manager devices.Manager) error {
	dm.Managers[manager.Name()] = manager
	return nil
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

func (dm *DeviceManager) Devices() ([]devices.Device, error) {
	var devices []devices.Device
	for _, m := range dm.Managers { 
		d, err := m.GetDevices()
		if err != nil {
			dm.log.Errorf("failed: %v", err)
		}
		devices = append(devices, d...)
	}
	return devices, nil
}

func (dm *DeviceManager) Start(dev devices.Device) error {
	for _, m := range dm.Managers {
		if m.HasDevice(dev) {
			return m.StartDevice(dev.DeviceID())
		}
	}
	return devices.DeviceNotFoundError
}

func (dm *DeviceManager) Stop(dev devices.Device) error {
	for _, m := range dm.Managers {
		if m.HasDevice(dev) {
			return m.StopDevice(dev.DeviceID())
		}
	}
	return devices.DeviceNotFoundError
}

func (dm *DeviceManager) isLocked(dev devices.Device) bool {
	for _, v := range dm.LockedDevices {
		if v.Device == dev {
			return true
		}
	}
	return false
}

func evaluateDevice(dev devices.Device, properties *DeviceProperties) bool {
	if properties == nil {
		return true
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

func (dm *DeviceManager) LockDevice(session *Session, properties *DeviceProperties) error {
	for _, manager := range dm.Managers {
		devices, _ := manager.GetDevices()
		for i := range devices {

			metsRequirements := evaluateDevice(devices[i], properties)

			if metsRequirements && !dm.isLocked(devices[i]) {
				dm.log.Debugf("LockDevice %v for session %s", devices[i], session.SessionID)
				lock := &DeviceLock{
					Device: devices[i],
				}
				dm.LockedDevices[session.SessionID] = lock
				session.Lock = lock
				return nil
			}
		}
	}
	return devices.ManagerNotFoundError
}

func (dm *DeviceManager) UnlockDevice(session *Session) error {

	if d, ok := dm.LockedDevices[session.SessionID]; ok {
		dm.log.Debugf("UnockDevice %v from session %s", d, session.SessionID)
		if d.Connection != nil {
			dm.log.Debugf("Cose Connection for device %v from session %s", d, session.SessionID)
			d.Connection.Close()
		}
		d.Device.StopApp(session.Properties.App, session.Properties.AppId)

		delete(dm.LockedDevices, session.SessionID)
		return nil
	}
	return fmt.Errorf("no locked device found for session")
}

func (dm *DeviceManager) Run(ctx context.Context) error {
	dm.log.Debug("Starting device manager")

	for _, v := range dm.Managers {
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

var (
	DeviceHandshakeRegex = regexp.MustCompile(`.*DEVICE_ID:([a-zA-Z0-9\-]+).*`)
	SessionHandshakeRegex = regexp.MustCompile(`.*SESSION_ID:([a-zA-Z0-9\-]+).*`)
)

func (dm *DeviceManager) lookupDevice(session *action.Session, remoteAddress string) *DeviceLock {

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
		lock.Device.SetConnectionState(devices.Connected)
		lock.ResponseChannel = make(chan []byte, 1)
		go handleMessages(dm.log, lock)
	}
}

func handleMessages(log *logrus.Entry, dev *DeviceLock) {
	for {
		chunkBuffer := make([]byte, 4)
		_, err := dev.Connection.Read(chunkBuffer)
		if err != nil {
			if io.EOF != err {
				log.Errorf("handleMessages ReadError: %v", err)
			} else {
				log.Info("Device connection closed")
			}
			dev.Connection.Close()
			close(dev.ResponseChannel)
			dev.Device.SetConnectionState(devices.Disconnected)
			dev.Connection = nil
			dev.WaitingGroup = nil
			return
		}

		r := bytes.NewReader(chunkBuffer)
		var messageSize uint32
		binary.Read(r, binary.LittleEndian, &messageSize)
		buffer := make([]byte, 0, messageSize)
		chunkBuffer = make([]byte, ReceiveBufferSize)
		for uint32(len(buffer)) < messageSize {
			n, err := dev.Connection.Read(chunkBuffer)
			if err != nil {
				log.Errorf("Chunk ReadError: %v", err)
			}
			buffer = append(buffer, chunkBuffer[:n]...)
		}

		dev.ResponseChannel <- buffer
		if dev.WaitingGroup != nil {
			dev.WaitingGroup.Done()
		}
	}
}

func (dm *DeviceManager) StopObserver() {
	dm.stop = true
}

func (dm *DeviceManager) SendBytes(session *Session, content []byte) ([]byte, error) {
	start := time.Now()
	var size []byte
	size = make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(content)))
	defer func() {
		elapsed := time.Since(start)
		session.logger.WithField("prefix", "dm").Debugf("send message took %s content: %s", elapsed.String(), content)
	}()

	if session.Lock.Connection != nil {
		session.Lock.WaitingGroup = &sync.WaitGroup{}
		session.Lock.WaitingGroup.Add(1)
		session.Lock.Connection.Write(size)
		session.Lock.Connection.Write(content)
		session.Lock.WaitingGroup.Wait()
		session.Lock.WaitingGroup = nil
		content := <-session.Lock.ResponseChannel
		return content, nil
	}

	return []byte{}, fmt.Errorf("device not connected")
}

func (dm *DeviceManager) Send(session *Session, content string) ([]byte, error) {
	return dm.SendBytes(session, []byte(content))
}

func (dm *DeviceManager)SendAction(logger *logrus.Entry, session *Session, act action.Interface) error {
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
	//logger.Debugf("Deserialize Action %v", response)
	return act.Deserialize(response)
}
