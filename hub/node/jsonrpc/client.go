// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package jsonrpc implements a JSON-RPC 1.0 ClientCodec and ServerCodec
// for the rpc package.
// For JSON-RPC 2.0 support, see https://godoc.org/?q=json-rpc+2.0
package jsonrpc

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/rpc"
	"sync"
)

type clientCodec struct {
	conn *websocket.Conn
	/*
		dec  *json.Decoder // for reading JSON values
		enc  *json.Encoder // for writing JSON values
	*/
	c io.Closer

	// temporary work space
	req  clientRequest
	resp clientResponse

	// JSON-RPC responses include the request id but not the request method.
	// Package rpc expects both.
	// We save the request method in pending when sending a request
	// and then look it up by request ID when filling out the rpc Response.
	mutex   sync.Mutex        // protects pending
	pending map[uint64]string // map request id to method name
}

// NewClientCodec returns a new [rpc.ClientCodec] using JSON-RPC on conn.
func NewClientCodec(conn *websocket.Conn) rpc.ClientCodec {
	return &clientCodec{
		conn: conn,
		/*
			dec:     json.NewDecoder(conn),
			enc:     json.NewEncoder(conn),
		*/
		c:       conn,
		pending: make(map[uint64]string),
	}
}

type clientRequest struct {
	Method string `json:"method"`
	Params [1]any `json:"params"`
	Id     uint64 `json:"id"`
}

func (c *clientCodec) WriteRequest(r *rpc.Request, param any) error {
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()
	c.req.Method = r.ServiceMethod
	c.req.Params[0] = param
	c.req.Id = r.Seq
	buffer, err := json.Marshal(c.req)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.BinaryMessage, buffer)
}

type clientResponse struct {
	Id     uint64           `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  any              `json:"error"`
}

func (r *clientResponse) reset() {
	r.Id = 0
	r.Result = nil
	r.Error = nil
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	c.resp.reset()
	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(msg, &c.resp); err != nil {
		return err
	}

	c.mutex.Lock()
	r.ServiceMethod = c.pending[c.resp.Id]
	delete(c.pending, c.resp.Id)
	c.mutex.Unlock()

	r.Error = ""
	r.Seq = c.resp.Id
	if c.resp.Error != nil || c.resp.Result == nil {
		x, ok := c.resp.Error.(string)
		if !ok {
			return fmt.Errorf("invalid error %v", c.resp.Error)
		}
		if x == "" {
			x = "unspecified error"
		}
		r.Error = x
	}
	return nil
}

func (c *clientCodec) ReadResponseBody(x any) error {
	if x == nil {
		return nil
	}
	return json.Unmarshal(*c.resp.Result, x)
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}

// NewClient returns a new [rpc.Client] to handle requests to the
// set of services at the other end of the connection.
func NewClient(conn *websocket.Conn) *rpc.Client {
	return rpc.NewClientWithCodec(NewClientCodec(conn))
}
