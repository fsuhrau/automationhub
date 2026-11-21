package videostream

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/videostream/videostreamtest"
	"github.com/gorilla/websocket"
)

type provider struct{ r *Rendezvous }

func (p provider) Rendezvous() *Rendezvous { return p.r }

func TestFrom(t *testing.T) {
	shared := NewRendezvous()
	if got := From(provider{shared}); got != shared {
		t.Error("From must return the rendezvous of the provider")
	}
	a, b := From(struct{}{}), From(nil)
	if a == nil || b == nil || a == b {
		t.Error("From without a provider must return a private rendezvous each time")
	}
}

func TestServeToMaster_DeliversStreamWithToken(t *testing.T) {
	r := NewRendezvous()
	tokens := make(chan string, 1)
	master := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, req, nil)
		if err != nil {
			return
		}
		tokens <- req.Header.Get("X-Auth-Token")
		if !r.Deliver(strings.TrimPrefix(req.URL.Path, "/node/video/"), conn) {
			_ = conn.Close()
		}
	}))
	defer master.Close()

	ch, done := r.Expect("s1")
	defer done()
	src := videostreamtest.New(device.VideoMeta{Codec: "h264"})
	target := MasterTarget{Host: strings.TrimPrefix(master.URL, "http://"), AuthToken: "secret"}
	if err := ServeToMaster(target, "s1", src); err != nil {
		t.Fatal(err)
	}
	stream, err := Await(t.Context(), ch, 2*time.Second)
	if err != nil || stream.Meta().Codec != "h264" {
		t.Fatalf("await: %v", err)
	}
	_ = stream.Close()
	if got := <-tokens; got != "secret" {
		t.Errorf("token = %q", got)
	}
}

func TestServeToMaster_NoTokenHeaderWithoutToken(t *testing.T) {
	tokens := make(chan []string, 1)
	master := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		tokens <- req.Header.Values("X-Auth-Token")
		if conn, err := (&websocket.Upgrader{}).Upgrade(w, req, nil); err == nil {
			_ = conn.Close()
		}
	}))
	defer master.Close()
	src := videostreamtest.New(device.VideoMeta{})
	if err := ServeToMaster(MasterTarget{Host: strings.TrimPrefix(master.URL, "http://")}, "s", src); err != nil {
		t.Fatal(err)
	}
	if got := <-tokens; len(got) != 0 {
		t.Errorf("unexpected token header %v", got)
	}
}

func TestServeToMaster_ClosesStreamWhenMasterUnreachable(t *testing.T) {
	src := videostreamtest.New(device.VideoMeta{})
	if err := ServeToMaster(MasterTarget{Host: "127.0.0.1:1"}, "s", src); err == nil {
		t.Fatal("expected a dial error")
	}
	select {
	case <-src.Closed():
	case <-time.After(time.Second):
		t.Fatal("stream not closed")
	}
}
