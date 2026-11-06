package videostream

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Rendezvous pairs the master's request for a node stream with the websocket
// the node opens for it. The master can't connect to a node, so the node
// dials back to /node/video/:stream_id.
type Rendezvous struct {
	mu      sync.Mutex
	waiting map[string]chan *websocket.Conn
}

func NewRendezvous() *Rendezvous {
	return &Rendezvous{waiting: make(map[string]chan *websocket.Conn)}
}

// Default is the rendezvous of the master process.
var Default = NewRendezvous()

// Expect registers a stream id. Call the returned func when done waiting.
func (r *Rendezvous) Expect(id string) (<-chan *websocket.Conn, func()) {
	ch := make(chan *websocket.Conn, 1)
	r.mu.Lock()
	r.waiting[id] = ch
	r.mu.Unlock()
	return ch, func() {
		r.mu.Lock()
		if r.waiting[id] == ch {
			delete(r.waiting, id)
		}
		r.mu.Unlock()
	}
}

// Deliver hands the node's websocket to the waiting party. It returns false
// when nobody waits for the id; the caller closes the socket then.
func (r *Rendezvous) Deliver(id string, conn *websocket.Conn) bool {
	r.mu.Lock()
	ch, ok := r.waiting[id]
	delete(r.waiting, id)
	r.mu.Unlock()
	if !ok {
		return false
	}
	ch <- conn
	return true
}

// Await waits for the node's websocket and reads the stream meta.
func Await(ctx context.Context, ch <-chan *websocket.Conn, timeout time.Duration) (device.VideoStream, error) {
	select {
	case conn := <-ch:
		return Receive(conn, timeout)
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, fmt.Errorf("node did not connect the video stream within %s", timeout)
	}
}

// ServeToMaster runs on the node: it dials the master's stream endpoint and
// pumps the stream into it in the background. The stream is closed when the
// master hangs up or the stream ends.
func ServeToMaster(masterHost, authToken, streamID string, stream device.VideoStream) error {
	u := url.URL{Scheme: "ws", Host: masterHost, Path: "/node/video/" + url.PathEscape(streamID)}
	headers := http.Header{}
	if authToken != "" {
		headers.Set("X-Auth-Token", authToken)
	}
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second, WriteBufferSize: 256 << 10}
	conn, _, err := dialer.Dial(u.String(), headers)
	if err != nil {
		_ = stream.Close()
		return fmt.Errorf("connect video stream to master: %w", err)
	}
	go func() {
		defer conn.Close()
		if err := Pump(context.Background(), conn, stream, InjectInto(stream)); err != nil {
			logrus.Debugf("video stream %s ended: %v", streamID, err)
		}
	}()
	return nil
}
