package videostream

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/gorilla/websocket"
)

// pump is one running Pump: it owns the writes to the socket.
type pump struct {
	conn   *websocket.Conn
	stream device.VideoStream
	handle InputHandler
	cancel context.CancelFunc

	writeMu sync.Mutex
}

// Feed is what Pump sends to a socket: a stream, and the handler of the input
// messages that come back (nil to ignore them).
type Feed struct {
	Stream device.VideoStream
	Input  InputHandler
}

// Pump writes the meta and then every packet of the stream to the websocket
// until the stream ends, a write fails or ctx is done. It closes the stream
// but not the socket. A reader goroutine passes the client's text messages to
// feed.Input (which may be nil) and notices close frames and disconnects; those
// end the pump as well.
func Pump(ctx context.Context, conn *websocket.Conn, feed Feed) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	p := &pump{conn: conn, stream: feed.Stream, handle: feed.Input, cancel: cancel}

	go p.readClient()
	// a blocked ReadPacket only returns once the stream is closed
	go func() {
		<-ctx.Done()
		_ = feed.Stream.Close()
	}()

	meta := feed.Stream.Meta()
	if err := p.send(Message{Type: TypeMeta, Meta: &meta}); err != nil {
		return err
	}
	return p.writePackets(ctx)
}

// send writes a control message, serialized with the packets.
func (p *pump) send(m Message) error {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	return writeJSON(p.conn, m)
}

// reply answers the client; a failed answer is dropped, the reader or the
// next packet write notices a broken socket.
func (p *pump) reply(m Message) { _ = p.send(m) }

// readClient reads the messages of the client until the socket ends, which
// ends the pump.
func (p *pump) readClient() {
	defer p.cancel()
	for {
		kind, data, err := p.conn.ReadMessage()
		if err != nil {
			return
		}
		p.handleMessage(kind, data)
	}
}

// handleMessage hands a text message of the client to the input handler.
func (p *pump) handleMessage(kind int, data []byte) {
	if kind != websocket.TextMessage || p.handle == nil {
		return
	}
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		p.reply(Message{Type: TypeInputError, Message: "invalid message: " + err.Error()})
		return
	}
	p.handle(m, p.reply)
}

// writePackets copies the packets of the stream to the socket; a stream that
// ended because ctx is done is no error.
func (p *pump) writePackets(ctx context.Context) error {
	for {
		packet, err := p.stream.ReadPacket()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if err := p.writePacket(packet); err != nil {
			return err
		}
	}
}

func (p *pump) writePacket(packet device.VideoPacket) error {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	_ = p.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return p.conn.WriteMessage(websocket.BinaryMessage, EncodePacket(packet))
}
