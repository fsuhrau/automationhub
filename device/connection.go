package device

import (
	"context"
	"errors"
	"fmt"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

var (
	DeviceDisconnectedError = fmt.Errorf("device disconnected")
)

const (
	DefaultSocketTimeout = 2 * time.Minute
	ReceiveBufferSize    = 20 * 1024
)

type ResponseData struct {
	Data []byte
	Err  error
}

type Connection struct {
	Logger              *logrus.Entry
	Connection          *websocket.Conn
	ResponseChannel     chan ResponseData
	ActionChannel       chan action.Response
	ConnectionParameter *action.Connect
}

func (c *Connection) HandleMessages(ctx context.Context) {
	defer func() {
		c.Logger.Info("HandleMessages finished")
		if err := recover(); err != nil {
			c.Logger.Error(err)
		}
	}()

	for {
		if err := c.Connection.SetReadDeadline(time.Now().Add(DefaultSocketTimeout)); err != nil {
			c.Logger.Errorf("SocketAccept SetDeadline: %v", err)
			return
		}
		time.Sleep(50 * time.Millisecond)
		if c.Connection == nil {
			return
		}
		_, data, err := c.Connection.ReadMessage()
		//fmt.Printf("data = %d - %v - %v\n", t, data, err)
		if err != nil {
			c.handleReadError(err)
			return
		}

		c.ResponseChannel <- ResponseData{Data: data, Err: nil}
	}
}

func (c *Connection) handleReadError(err error) {
	if !errors.Is(err, io.ErrUnexpectedEOF) && io.EOF != err {
		c.Logger.Infof("Device disconnected: %v", err)
	} else {
		c.Logger.Info("Device disconnected")
	}
	if c.ResponseChannel != nil {
		c.ResponseChannel <- ResponseData{Data: nil, Err: DeviceDisconnectedError}
	}
}

func (c *Connection) Close() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	if c.Connection != nil {
		_ = c.Connection.Close()
	}
	c.Connection = nil

	close(c.ResponseChannel)
	for range c.ResponseChannel {
	}
	close(c.ActionChannel)
	for range c.ActionChannel {
	}
}

func (c *Connection) Send(content []byte) error {
	if c == nil || c.Connection == nil {
		return fmt.Errorf("device not connected")
	}
	if err := c.Connection.WriteMessage(websocket.TextMessage, content); err != nil {
		return err
	}
	return nil
}
