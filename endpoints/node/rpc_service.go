package node

import (
	"fmt"
	node2 "github.com/fsuhrau/automationhub/hub/node"
	"github.com/gorilla/websocket"
	"github.com/matishsiao/goInfo"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"os"
)

var rpcServer = rpc.DefaultServer

func (s *Service) ConnectAndServe() {

	serverURL := url.URL{Scheme: "ws", Host: s.cfg.MasterURL, Path: "/node/connect"}

	c, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
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
