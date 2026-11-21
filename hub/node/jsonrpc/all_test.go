package jsonrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/rpc"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

type Args struct {
	A, B int
}

type Reply struct {
	C int
}

type Arith int

func (t *Arith) Add(args *Args, reply *Reply) error {
	reply.C = args.A + args.B
	return nil
}

func (t *Arith) Div(args *Args, reply *Reply) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	reply.C = args.A / args.B
	return nil
}

type BlobRequest struct {
	Size int
}

type BlobReply struct {
	Data []byte
}

func (t *Arith) Blob(args *BlobRequest, reply *BlobReply) error {
	reply.Data = bytes.Repeat([]byte{0xAB}, args.Size)
	return nil
}

// newPair starts a websocket server serving the rpc server and returns a
// connected rpc client.
func newPair(t *testing.T) *rpc.Client {
	t.Helper()
	return NewClient(dialServer(t))
}

// dialServer starts a websocket server serving the rpc server and returns
// the client side of its socket.
func dialServer(t *testing.T) *websocket.Conn {
	t.Helper()
	server := rpc.NewServer()
	if err := server.Register(new(Arith)); err != nil {
		t.Fatal(err)
	}
	upgrader := websocket.Upgrader{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		server.ServeCodec(NewServerCodec(conn))
	}))
	t.Cleanup(ts.Close)

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestCall(t *testing.T) {
	client := newPair(t)
	var reply Reply
	if err := client.Call("Arith.Add", &Args{A: 7, B: 8}, &reply); err != nil {
		t.Fatal(err)
	}
	if reply.C != 15 {
		t.Fatalf("expected 15, got %d", reply.C)
	}
}

func TestServerError(t *testing.T) {
	client := newPair(t)
	var reply Reply
	err := client.Call("Arith.Div", &Args{A: 1, B: 0}, &reply)
	if err == nil || err.Error() != "divide by zero" {
		t.Fatalf("expected divide by zero, got %v", err)
	}
}

func TestUnknownMethod(t *testing.T) {
	client := newPair(t)
	var reply Reply
	if err := client.Call("Arith.Unknown", &Args{}, &reply); err == nil {
		t.Fatal("expected error for unknown method")
	}
}

func TestConcurrentCalls(t *testing.T) {
	client := newPair(t)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply Reply
			if err := client.Call("Arith.Add", &Args{A: i, B: i}, &reply); err != nil {
				t.Error(err)
				return
			}
			if reply.C != 2*i {
				t.Errorf("expected %d, got %d", 2*i, reply.C)
			}
		}(i)
	}
	wg.Wait()
}

// screenshots are transferred in chunks of this size; make sure they survive
// the JSON (base64) encoding over a single websocket message.
func TestLargePayload(t *testing.T) {
	client := newPair(t)
	var reply BlobReply
	const size = 512 * 1024
	if err := client.Call("Arith.Blob", &BlobRequest{Size: size}, &reply); err != nil {
		t.Fatal(err)
	}
	if len(reply.Data) != size {
		t.Fatalf("expected %d bytes, got %d", size, len(reply.Data))
	}
}

// A request without an id is answered with the id null.
func TestRequestWithoutIdIsAnsweredWithNull(t *testing.T) {
	conn := dialServer(t)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"method":"Arith.Add","params":[{"A":1,"B":2}]}`)); err != nil {
		t.Fatal(err)
	}

	var reply map[string]json.RawMessage
	if err := conn.ReadJSON(&reply); err != nil {
		t.Fatal(err)
	}

	if id, ok := reply["id"]; !ok || string(id) != "null" || !strings.Contains(string(reply["result"]), `"C":3`) {
		t.Fatalf("reply %s", reply)
	}
}
