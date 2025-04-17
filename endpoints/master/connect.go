package master

import (
	"fmt"
	"github.com/fsuhrau/automationhub/config/protocol"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  protocol.SocketFrameSize,
	WriteBufferSize: protocol.SocketFrameSize,
}

func (s *nodeMaster) Connect(c *gin.Context) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	fmt.Println("Update Connection")
	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v\n", err)
		return
	}

	go s.handleNode(conn, c)
}

func (s *nodeMaster) handleNode(conn *websocket.Conn, c *gin.Context) {

	// read connect message first
	_, data, err := conn.ReadMessage()
	if err != nil {
		fmt.Printf("Failed to read message: %+v\n", err)
		return
	}

	var request node.RegisterNodeRequest
	if err := proto.Unmarshal(data, &request); err != nil {
		return
	}

	fmt.Printf("node registered: %s\n", request.Hostname)

	rpcClient := node.NewRPCClient(conn, s.cfg.MasterURL)

	s.nm.RegisterNode(manager.NodeIdentifier(request.GetIdentifier()), rpcClient, request.GetHostname(), c.RemoteIP(), request.GetOperationSystem(), request.GetEnvironmentVariables(), request.GetPort(), request.GetManagers())

	events.NodeConnected.Trigger(events.NodeConnectedPayload{
		NodeIdentifier: manager.NodeIdentifier(request.GetIdentifier()),
		Hostname:       request.GetHostname(),
	})
}
