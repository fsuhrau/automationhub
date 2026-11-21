package hub

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (dm *DeviceManager) RegisterRoutes(r *gin.Engine, auth *gin.RouterGroup) error {
	deviceApi := auth.Group("/device")
	deviceApi.GET("/connect", func(c *gin.Context) {
		conn, err := dm.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			dm.log.Errorf("Failed to set websocket upgrade: %+v", err)
			return
		}

		dm.handleConnection(conn)
	})

	return nil
}

func (dm *DeviceManager) handleConnection(conn *websocket.Conn) {
	if err := conn.SetReadDeadline(time.Now().Add(DeviceConnectionTimeout)); err != nil {
		dm.log.Errorf("WebSocketAccept SetDeadline: %v", err)
	}

	remoteAddress := conn.RemoteAddr().String()
	dm.log.Infof("WebSocketAccept established: %v", remoteAddress)

	connect, err := readHandshake(conn)
	if err != nil {
		dm.log.Errorf("WebSocketAccept: %+v", err)
		return
	}

	dev, _ := dm.GetDevice(connect.DeviceID)
	if dev == nil {
		// without this the app waits and the test run times out with no hint
		dm.log.Warnf("WebSocketAccept: %v sent a handshake for unknown device %q (session %q); closing", remoteAddress, connect.DeviceID, connect.SessionID)
		_ = conn.Close()
		return
	}
	dm.log.Infof("Received Handshake from %v", remoteAddress)
	dm.log.Debugf("Device with ID %s connected", dev.DeviceID())
	dm.attachConnection(dev, conn, connect)
}

func readHandshake(conn *websocket.Conn) (*action.Connect, error) {
	_, buffer, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	resp := &action.Response{}
	if err := json.Unmarshal(buffer, resp); err != nil {
		return nil, fmt.Errorf("handshake: %w", err)
	}
	if resp.Payload.Connect == nil {
		return nil, fmt.Errorf("handshake without connect payload")
	}
	return resp.Payload.Connect, nil
}

func (dm *DeviceManager) attachConnection(dev device.Device, conn *websocket.Conn, connect *action.Connect) {
	connection := &device.Connection{
		ConnectionParameter: connect,
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
