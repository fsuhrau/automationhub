package master

import (
	"fmt"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func (s *nodeMaster) Connect(c *gin.Context) {
	conn, err := s.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
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

// Video accepts the websocket a node opens to deliver a device video stream
// the master asked for (see videostream.Rendezvous). The rendezvous is the one
// of the node manager, which the node devices wait on.
func (s *nodeMaster) Video(c *gin.Context) {
	conn, err := s.videoUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	if !videostream.From(s.nm).Deliver(c.Param("stream_id"), conn) {
		_ = conn.Close()
	}
}
