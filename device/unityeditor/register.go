package unityeditor

import (
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// RegisterDevice adds an editor that connected, or refreshes the known one,
// and serves its websocket in the background.
func (m *Handler) RegisterDevice(data device.RegisterData) (device.Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.clock.Now().UTC()
	dev, known := m.Get(data.DeviceID)
	if known {
		dev.applyRegistration(data, now)
		m.storeUpdate(dev)
	} else {
		dev = m.registerNew(data, now)
	}
	go dev.HandleManagerConnection()
	return dev, nil
}

func (m *Handler) registerNew(data device.RegisterData, now time.Time) *Device {
	dev := newDevice(generic.Info{ID: data.DeviceID}, m.clock)
	dev.applyRegistration(data, now)
	model := newModel(data)
	dev.SetConfig(&model)
	if err := m.deviceStorage.NewDevice(Manager, model); err != nil {
		logrus.WithError(err).Warnf("unity editor %s: store new device", data.DeviceID)
	}
	m.Add(dev)
	return dev
}

func editorName(data device.RegisterData) string {
	return fmt.Sprintf("Unity (%s) @ %s", data.TargetVersion, data.Name)
}

// applyRegistration copies what the editor reported into the device.
func (d *Device) applyRegistration(data device.RegisterData, now time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.OSName = data.DeviceOS
	d.OSVersion = data.DeviceOSVersion
	d.Target = data.TargetVersion
	d.Name = editorName(data)
	d.osInfos = data.DeviceOSInfos
	d.ip = data.DeviceIP
	d.lastUpdateAt = now
	d.conn = connOf(data.Conn)
	d.updated = true
}

// connOf keeps a missing websocket a nil interface.
func connOf(conn *websocket.Conn) managerConn {
	if conn == nil {
		return nil
	}
	return conn
}

func newModel(data device.RegisterData) models.Device {
	return models.Device{
		DeviceIdentifier: data.DeviceID,
		DeviceType:       models.DeviceTypeUnityEditor,
		Name:             editorName(data),
		Manager:          Manager,
		OS:               data.DeviceOS,
		OSVersion:        data.DeviceOSVersion,
		OSInfos:          data.DeviceOSInfos,
		TargetVersion:    data.TargetVersion,
		ConnectionParameter: &models.ConnectionParameter{
			ConnectionType: models.ConnectionTypeRemote,
		},
		CustomParameter: []models.CustomParameter{
			{Key: "projectDir", Value: data.ProjectDir},
		},
	}
}
