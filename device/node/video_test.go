package node

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/fsuhrau/automationhub/hub/videostream/videostreamtest"
	"github.com/gorilla/websocket"
)

// videoNodes plays the node side: StartVideoStream dials back to the
// master's video endpoint like RPCNode.StartVideoStream does.
type videoNodes struct {
	manager.Nodes
	masterHost string
	stream     *videostreamtest.Stream
	req        device.VideoStreamRequest
}

func (n *videoNodes) StartVideoStream(_ manager.NodeIdentifier, req device.VideoStreamRequest) error {
	n.req = req
	return videostream.ServeToMaster(n.masterHost, "", req.StreamID, n.stream)
}

func TestNodeDeviceVideoRelay(t *testing.T) {
	// the master's /node/video/:stream_id endpoint
	master := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		if !videostream.Default.Deliver(strings.TrimPrefix(r.URL.Path, "/node/video/"), conn) {
			_ = conn.Close()
		}
	}))
	defer master.Close()

	meta := device.VideoMeta{Codec: "h264", Width: 720, Height: 1280}
	src := videostreamtest.New(meta, device.VideoPacket{KeyFrame: true, PTS: 5, Data: []byte{0, 0, 0, 1, 0x65}})
	nodes := &videoNodes{masterHost: strings.TrimPrefix(master.URL, "http://"), stream: src}
	dev := &NodeDevice{nodeId: "node-1", nodeManager: nodes, deviceID: "emulator-5554"}

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := dev.StartVideo(ctx, device.VideoOptions{MaxSize: 1280})
	if err != nil {
		t.Fatal(err)
	}
	if nodes.req.DeviceID != "emulator-5554" || nodes.req.Options.MaxSize != 1280 || nodes.req.StreamID == "" {
		t.Fatalf("unexpected request %+v", nodes.req)
	}
	if stream.Meta() != meta {
		t.Fatalf("meta %+v", stream.Meta())
	}
	p, err := stream.ReadPacket()
	if err != nil || !p.KeyFrame || p.PTS != 5 {
		t.Fatalf("packet %+v %v", p, err)
	}

	// input travels to the node's stream
	key := device.InputEvent{Kind: device.InputKey, Action: device.InputPress, Key: "back"}
	if err := stream.(device.VideoInput).InjectInput(key); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-src.Inputs:
		if got != key {
			t.Fatalf("input %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("input not relayed to the node")
	}

	// the viewer leaves: the node side stream is closed too
	cancel()
	select {
	case <-src.Closed():
	case <-time.After(2 * time.Second):
		t.Fatal("node stream not closed")
	}
}
