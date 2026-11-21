package androiddevice

import (
	"context"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
)

func (m *Handler) startADBServer() error {
	_, err := m.adbHost("start-server")
	return err
}

// Start starts the adb server and subscribes to its device events, so
// devices that are plugged in, unplugged or authorized show up right away.
func (m *Handler) Start() error {
	err := m.startADBServer()

	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.stopWatch = cancel
	m.mu.Unlock()
	w := adbWatcher{
		addr:        m.adbServerAddr,
		clock:       m.deps.Clock,
		startServer: m.startADBServer,
		onList: func(entries []adbDeviceEntry) {
			m.applyDeviceList(entries)
			m.changed()
		},
		connected: m.tracking.Store,
	}
	go w.run(ctx)
	return err
}

func (m *Handler) Stop() error {
	m.mu.Lock()
	if m.stopWatch != nil {
		m.stopWatch()
		m.stopWatch = nil
	}
	m.mu.Unlock()
	_, err := m.adbHost("kill-server")
	return err
}

// StartDevice reconnects a network device that adb doesn't list. An unplugged
// USB device can't be started from here.
func (m *Handler) StartDevice(deviceID string) error {
	d, ok := m.Get(deviceID)
	if !ok {
		return device.DeviceNotFoundError
	}
	if d.DeviceState() != device.StateRemoteDisconnected {
		return nil
	}
	cfg, err := m.deviceStorage.GetDevice(Manager, deviceID)
	if err != nil {
		return device.DeviceNotFoundError
	}
	return m.adbRemote("connect", cfg)
}

// StopDevice disconnects a network device.
func (m *Handler) StopDevice(deviceID string) error {
	d, ok := m.Get(deviceID)
	if !ok {
		return device.DeviceNotFoundError
	}
	return m.adbRemote("disconnect", d.GetConfig())
}

// registerNewDevice stores a newly seen device and wires its remote logging.
func (m *Handler) registerNewDevice(d *Device) {
	cfg := models.Device{
		DeviceIdentifier: d.ID,
		DeviceType:       models.DeviceTypePhone,
		Name:             d.Name,
		Manager:          Manager,
		OS:               "android",
		PlatformType:     models.PlatformTypeAndroid,
		OSVersion:        d.OSVersion,
		ConnectionParameter: &models.ConnectionParameter{
			ConnectionType: models.ConnectionTypeUSB,
		},
	}
	m.wireRemote(d)
	if err := m.deviceStorage.NewDevice(m.Name(), cfg); err != nil {
		logrus.Warnf("create android device %s: %v", d.ID, err)
	}
	m.persist(d)
	d.SetConfig(&cfg)
}
