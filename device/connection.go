package device

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

var (
	DeviceDisconnectedError = fmt.Errorf("device disconnected")
)

const (
	DefaultSocketTimeout = 60 * time.Minute
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
}

func GetMessageSize(buffer []byte) uint32 {
	r := bytes.NewReader(buffer)
	var messageSize uint32
	_ = binary.Read(r, binary.LittleEndian, &messageSize)
	return messageSize
}

func (c *Connection) HandleMessages() {
	defer func() {
		c.Logger.Info("HandleMessages finished")
		if err := recover(); err != nil {
			c.Logger.Error(err)
		}
	}()

	sizeBuffer := make([]byte, 4)
	chunkBuffer := make([]byte, ReceiveBufferSize)
	for {
		if err := c.Connection.SetDeadline(time.Now().Add(DefaultSocketTimeout)); err != nil {
			c.Logger.Errorf("SocketAccept SetDeadline: %v", err)
			return
		}

		_, err := c.Connection.Read(sizeBuffer)
		if err != nil {
			c.handleReadError(err)
			return
		}
		messageSize := GetMessageSize(sizeBuffer)

		var responseData ResponseData
		responseData.Data = make([]byte, 0, messageSize)
		for uint32(len(responseData.Data)) < messageSize {
			n, err := c.Connection.Read(chunkBuffer)
			if err != nil {
				responseData.Err = err
				c.Logger.Errorf("Chunk ReadError: %v", err)
				break
			}
			responseData.Data = append(responseData.Data, chunkBuffer[:n]...)
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

	if c.Connection != nil {
		c.Connection.Close()
	}
	if c.ResponseChannel != nil {
		c.ResponseChannel <- responseData
	}
	close(c.ResponseChannel)
	c.Connection = nil
	c.ConnectionStateChannel <- Disconnected
}

func (c *Connection) Close() {
	if c.Connection != nil {
		_ = c.Connection.Close()
	}
	close(c.ConnectionStateChannel)
}

func (c *Connection) Send(content []byte) error {
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
