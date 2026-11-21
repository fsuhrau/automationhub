package node

import (
	"errors"
	netrpc "net/rpc"
	"time"

	"github.com/fsuhrau/automationhub/config/protocol"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/node/jsonrpc"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// rpcService is the name the node registers its RPC receiver under; it is
// part of the wire format.
const rpcService = "RPCNode."

// RPCClient is the master's RPC connection to one node.
type RPCClient struct {
	client    *netrpc.Client
	masterURL string

	uploadInterval time.Duration
	uploadStall    time.Duration
}

func NewRPCClient(conn *websocket.Conn, masterURL string) *RPCClient {
	conn.SetReadLimit(protocol.SocketFrameSize)
	return &RPCClient{
		client:         jsonrpc.NewClient(conn),
		masterURL:      masterURL,
		uploadInterval: time.Second,
		uploadStall:    UploadStallTimeout,
	}
}

func (c *RPCClient) safeCall(serviceMethod string, args interface{}, reply interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Recovered from panic in %s: %v", serviceMethod, r)
		}
	}()
	return c.client.Call(serviceMethod, args, reply)
}

// failing is a node response that can carry an error the node reported.
type failing[R any] interface {
	*R
	failure() error
}

// call sends one request and returns the node's response.
func call[R any](c *RPCClient, method string, req any) (R, error) {
	logrus.Info(rpcService + method)
	var resp R
	err := c.safeCall(rpcService+method, req, &resp)
	return resp, err
}

// forward is call for responses that carry an error code: a failure the node
// reported becomes the returned error. Every forwarded call is one line.
func forward[R any, P failing[R]](c *RPCClient, method string, req any) (R, error) {
	resp, err := call[R](c, method, req)
	if err != nil {
		return resp, err
	}
	return resp, P(&resp).failure()
}

// forwardOK forwards a call whose BoolResponse only matters for its error.
func (c *RPCClient) forwardOK(method string, req any) error {
	_, err := forward[BoolResponse](c, method, req)
	return err
}

// forwardBool forwards a call and returns the value of its BoolResponse.
func (c *RPCClient) forwardBool(method string, req any) (bool, error) {
	resp, err := forward[BoolResponse](c, method, req)
	if err != nil {
		return false, err
	}
	return resp.Value, nil
}

// notify forwards a call whose outcome the caller doesn't care about.
func (c *RPCClient) notify(method string, req any) {
	_, _ = call[Void](c, method, req)
}

// controlError maps an error the node returned to a device.ControlError.
func controlError(err error) error {
	var serverErr netrpc.ServerError
	if errors.As(err, &serverErr) {
		return device.ControlError(string(serverErr))
	}
	return err
}

func (c *RPCClient) Ping() error {
	return c.safeCall(rpcService+"Ping", &Void{}, &Void{})
}

func (c *RPCClient) GetDevices() (map[string][]device.Device, error) {
	resp, err := call[DevicesResponse](c, "GetDevices", &Void{})
	if err != nil {
		return nil, err
	}
	return FromDevicesResponse(resp.Devices), nil
}

func (c *RPCClient) StartDevice(deviceId string) error {
	_, err := forward[ErrorResponse](c, "StartDevice", &DeviceRequest{DeviceID: deviceId})
	return err
}

func (c *RPCClient) StopDevice(deviceId string) error {
	_, err := forward[ErrorResponse](c, "StopDevice", &DeviceRequest{DeviceID: deviceId})
	return err
}
