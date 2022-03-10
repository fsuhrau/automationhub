package device

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

var (
	DeviceDisconnectedError = fmt.Errorf("device disconnected")
)

const (
	DefaultSocketTimeout = 1 * time.Minute
	ReceiveBufferSize    = 20 * 1024
)

type ResponseData struct {
	Data []byte
	Err  error
}

type Connection struct {
	Logger                 *logrus.Entry
	Connection             net.Conn
	ResponseChannel        chan ResponseData
	ConnectionStateChannel chan ConnectionState
	ActionChannel          chan action.Response
	ConnectionParameter    *action.Connect
}

func GetMessageSize(buffer []byte) uint32 {
	r := bytes.NewReader(buffer)
	var messageSize uint32
	_ = binary.Read(r, binary.LittleEndian, &messageSize)
	return messageSize
}

func (c *Connection) HandleMessages(ctx context.Context) {
	defer func() {
		c.Logger.Info("HandleMessages finished")
		if err := recover(); err != nil {
			c.Logger.Error(err)
		}
	}()

	messageSizeBuf := make([]byte, 4)
	r := bufio.NewReaderSize(c.Connection, ReceiveBufferSize)
	for {
		if err := c.Connection.SetDeadline(time.Now().Add(DefaultSocketTimeout)); err != nil {
			c.Logger.Errorf("SocketAccept SetDeadline: %v", err)
			return
		}

		_, err := io.ReadFull(r, messageSizeBuf[:4])
		if err != nil {
			c.handleReadError(err)
			return
		}

		messageSize := GetMessageSize(messageSizeBuf)

		var responseData ResponseData
		responseData.Data = make([]byte, messageSize)

		_, err = io.ReadFull(r, responseData.Data[:messageSize])
		if err != nil {
			responseData.Err = err
			c.handleReadError(err)
			break
		}
		c.ResponseChannel <- responseData
	}
}

func (c *Connection) handleReadError(err error) {
	if io.EOF != err {
		c.Logger.Info("Device disconnected: %v", err)
	} else {
		c.Logger.Info("Device disconnected")
	}
	var responseData ResponseData
	responseData.Err = DeviceDisconnectedError
	if c.ResponseChannel != nil {
		c.ResponseChannel <- responseData
	}
	c.Close()
}

func (c *Connection) Close() {

	if c.ResponseChannel != nil {
		close(c.ResponseChannel)
	}

	if c.Connection != nil {
		_ = c.Connection.Close()
	}
	c.Connection = nil
}

func (c *Connection) Send(content []byte) error {
	if c == nil || c.Connection == nil {
		return fmt.Errorf("device not connected")
	}
	var size []byte
	size = make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(content)))
	if _, err := c.Connection.Write(size); err != nil {
		return err
	}
	if _, err := c.Connection.Write(content); err != nil {
		return err
	}
	return nil
}
