package scrcpy

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"testing"
)

func header(codec uint32, w, h uint32) []byte {
	b := make([]byte, 12)
	binary.BigEndian.PutUint32(b, codec)
	binary.BigEndian.PutUint32(b[4:], w)
	binary.BigEndian.PutUint32(b[8:], h)
	return b
}

func packet(ptsFlags uint64, data []byte) []byte {
	b := make([]byte, 12, 12+len(data))
	binary.BigEndian.PutUint64(b, ptsFlags)
	binary.BigEndian.PutUint32(b[8:], uint32(len(data)))
	return append(b, data...)
}

func TestReadHeader(t *testing.T) {
	codec, w, h, err := readHeader(bytes.NewReader(header(codecH264, 1080, 2400)))
	if err != nil || codec != "h264" || w != 1080 || h != 2400 {
		t.Fatalf("got %q %dx%d %v", codec, w, h, err)
	}
	if _, _, _, err := readHeader(bytes.NewReader([]byte{0, 0, 0, 1})); err == nil || !strings.Contains(err.Error(), "configuration error") {
		t.Fatalf("expected configuration error, got %v", err)
	}
	if _, _, _, err := readHeader(bytes.NewReader([]byte{0, 0, 0, 0})); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled error, got %v", err)
	}
}

func TestReadPacket(t *testing.T) {
	var stream bytes.Buffer
	stream.Write(packet(flagConfig, []byte{0, 0, 0, 1, 0x67}))
	stream.Write(packet(flagKeyFrame|123456, []byte{0, 0, 0, 1, 0x65, 1}))
	stream.Write(packet(123500, []byte{0, 0, 0, 1, 0x41}))

	p, err := readPacket(&stream)
	if err != nil || !p.Config || p.KeyFrame || p.PTS != 0 || len(p.Data) != 5 {
		t.Fatalf("config packet: %+v %v", p, err)
	}
	p, err = readPacket(&stream)
	if err != nil || p.Config || !p.KeyFrame || p.PTS != 123456 || len(p.Data) != 6 {
		t.Fatalf("key frame: %+v %v", p, err)
	}
	p, err = readPacket(&stream)
	if err != nil || p.Config || p.KeyFrame || p.PTS != 123500 {
		t.Fatalf("frame: %+v %v", p, err)
	}
	if _, err := readPacket(&stream); err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestReadPacketRejectsBadSize(t *testing.T) {
	b := make([]byte, 12)
	binary.BigEndian.PutUint32(b[8:], maxPacketSize+1)
	if _, err := readPacket(bytes.NewReader(b)); err == nil {
		t.Fatal("expected an error for an oversized packet")
	}
}

func TestServerArgs(t *testing.T) {
	args := strings.Join(serverArgs(0x1234, Options{MaxSize: 1280, MaxFPS: 30, BitRate: 4000000}), " ")
	for _, want := range []string{"scid=00001234", "audio=false", "control=false", "tunnel_forward=true", "max_size=1280", "max_fps=30", "video_bit_rate=4000000"} {
		if !strings.Contains(args, want) {
			t.Errorf("args %q miss %q", args, want)
		}
	}
}

func TestEmbeddedServer(t *testing.T) {
	// the jar is a zip file
	if len(serverJar) < 1024 || !bytes.HasPrefix(serverJar, []byte("PK")) {
		t.Fatalf("embedded scrcpy server looks wrong (%d bytes)", len(serverJar))
	}
}

func TestEncodeTouch(t *testing.T) {
	b := encodeTouch(ActionDown, 7, 100, 200, 576, 1280)
	want := []byte{
		typeInjectTouch, ActionDown,
		0, 0, 0, 0, 0, 0, 0, 7, // pointer id
		0, 0, 0, 100, 0, 0, 0, 200, // x, y
		0x02, 0x40, 0x05, 0x00, // 576, 1280
		0xff, 0xff, // pressure 1.0
		0, 0, 0, 0, 0, 0, 0, 0, // action button, buttons
	}
	if !bytes.Equal(b, want) {
		t.Fatalf("touch down\n got %x\nwant %x", b, want)
	}
	if up := encodeTouch(ActionUp, 7, 1, 1, 2, 2); up[22] != 0 || up[23] != 0 {
		t.Fatalf("touch up must have no pressure: %x", up)
	}
}

func TestEncodeKeyTextScroll(t *testing.T) {
	if b := encodeKeycode(ActionUp, 4); !bytes.Equal(b, []byte{0, 1, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0}) {
		t.Fatalf("keycode %x", b)
	}
	b, err := encodeText("hi")
	if err != nil || !bytes.Equal(b, []byte{typeInjectText, 0, 0, 0, 2, 'h', 'i'}) {
		t.Fatalf("text %x %v", b, err)
	}
	if _, err := encodeText(strings.Repeat("x", injectTextMaxLength+1)); err == nil {
		t.Fatal("expected an error for too long text")
	}
	s := encodeScroll(1, 2, 3, 4, 0, -16)
	if len(s) != 21 || s[0] != typeInjectScroll || binary.BigEndian.Uint16(s[13:]) != 0 || int16(binary.BigEndian.Uint16(s[15:])) != -0x8000 {
		t.Fatalf("scroll %x", s)
	}
	if v := i16FixedPoint(1); v != 0x7fff {
		t.Fatalf("fixed point 1.0 = %#x", v)
	}
}

func TestSessionWritesControl(t *testing.T) {
	if err := (&Session{}).InjectKeycode(ActionDown, 4); err != ErrNoControl {
		t.Fatalf("expected ErrNoControl, got %v", err)
	}
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	s := &Session{control: client}
	go func() { _ = s.InjectTouch(ActionMove, -1, 5, 6, 10, 20) }()
	buf := make([]byte, 32)
	if _, err := io.ReadFull(server, buf); err != nil {
		t.Fatal(err)
	}
	if buf[1] != ActionMove || int64(binary.BigEndian.Uint64(buf[2:])) != -2 {
		t.Fatalf("mouse pointer id must become the generic finger: %x", buf)
	}
}
