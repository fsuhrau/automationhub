package node

import (
	"fmt"
	"github.com/fsuhrau/automationhub/config/protocol"
	node2 "github.com/fsuhrau/automationhub/hub/node"
	"github.com/gorilla/websocket"
	"github.com/matishsiao/goInfo"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"os"
	"time"
)

var rpcServer = rpc.DefaultServer

func (s *Service) ConnectAndServe() {

	serverURL := url.URL{Scheme: "ws", Host: s.cfg.MasterURL, Path: "/node/connect"}

	proxyURL, _ := url.Parse("http://10.35.111.236:8888")

	dialer := websocket.Dialer{
		Proxy:             http.ProxyURL(proxyURL),
		HandshakeTimeout:  45 * time.Second,
		ReadBufferSize:    protocol.SocketFrameSize,
		WriteBufferSize:   protocol.SocketFrameSize,
		EnableCompression: true,
	}
	c, _, err := dialer.Dial(serverURL.String(), nil)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	s.c = c

	gi, _ := goInfo.GetInfo()
	if err := s.RegisterNode(s.cfg.Identifier, gi.Hostname, gi.OS, os.Getenv("PATH"), s.cfg.Port, s.managers); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	rpcServer.ServeCodec(jsonrpc.NewServerCodec(&node2.WebSocketConn{Conn: c}))
}
