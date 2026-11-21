package unityeditor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fsuhrau/automationhub/app"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/gorilla/websocket"
)

const sendQueueSize = 10

var errSendQueueFull = errors.New("editor send queue is full")

// managerConn is the websocket the editor registered with.
type managerConn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(messageType int, data []byte) error
}

type startRequest struct {
	Action    string
	HostUrl   string
	SessionID string
}

type stopRequest struct {
	Action string
}

//nolint:revive // argument-limit: the signature is fixed by device.AppManager
func (d *Device) StartApp(_ *device.DeviceConfig, _ *app.Parameter, sessionId string, nodeUrl string) error {
	return d.sendCommand(startRequest{Action: "start", SessionID: sessionId, HostUrl: nodeUrl})
}

func (d *Device) StopApp(_ *app.Parameter) error {
	return d.sendCommand(stopRequest{Action: "stop"})
}

// sendCommand queues the command for the editor's websocket.
func (d *Device) sendCommand(command any) error {
	buffer, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("encode editor command: %w", err)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.sendCh == nil {
		return generic.ErrNotConnected
	}
	select {
	case d.sendCh <- buffer:
		return nil
	default:
		return errSendQueueFull
	}
}

// HandleManagerConnection serves the editor's websocket until it closes. The
// device is booted while the connection lives.
func (d *Device) HandleManagerConnection() {
	conn := d.openConnection()
	if conn == nil {
		return
	}
	defer d.closeConnection()
	go d.writeLoop(conn)
	d.readLoop(conn)
}

func (d *Device) openConnection() managerConn {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.conn == nil {
		return nil
	}
	d.State = device.StateBooted
	d.updated = true
	d.sendCh = make(chan []byte, sendQueueSize)
	return d.conn
}

func (d *Device) closeConnection() {
	d.mu.Lock()
	defer d.mu.Unlock()
	close(d.sendCh)
	d.sendCh = nil
	d.State = device.StateShutdown
	d.updated = true
}

func (d *Device) writeLoop(conn managerConn) {
	d.mu.Lock()
	queue := d.sendCh
	d.mu.Unlock()
	for message := range queue {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

func (d *Device) readLoop(conn managerConn) {
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
		d.mu.Lock()
		d.lastUpdateAt = d.clock.Now().UTC()
		d.mu.Unlock()
	}
}
