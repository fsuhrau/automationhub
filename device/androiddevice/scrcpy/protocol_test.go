package scrcpy

import (
	"bytes"
	"io"
	"net"
	"strings"
	"testing"
)

func TestCodecName(t *testing.T) {
	tests := []struct {
		id      uint32
		want    string
		wantErr string
	}{
		{codecH264, "h264", ""},
		{codecH265, "h265", ""},
		{codecAV1, "av1", ""},
		{0, "", "disabled"},
		{1, "", "configuration error"},
		{0xdead, "", "unknown scrcpy codec id 0xdead"},
	}
	for _, tt := range tests {
		got, err := codecName(tt.id)
		if got != tt.want || (err == nil) != (tt.wantErr == "") || (err != nil && !strings.Contains(err.Error(), tt.wantErr)) {
			t.Errorf("codecName(%#x) = %q, %v", tt.id, got, err)
		}
	}
}

func TestReadHeaderShortInput(t *testing.T) {
	if _, _, _, err := readHeader(bytes.NewReader(nil)); err == nil || !strings.Contains(err.Error(), "read scrcpy codec") {
		t.Fatalf("codec: %v", err)
	}
	if _, _, _, err := readHeader(bytes.NewReader(header264(1, 1)[:6])); err == nil || !strings.Contains(err.Error(), "read scrcpy video size") {
		t.Fatalf("size: %v", err)
	}
	if _, _, _, err := readHeader(bytes.NewReader([]byte{0, 0, 0xde, 0xad})); err == nil || !strings.Contains(err.Error(), "unknown scrcpy codec") {
		t.Fatalf("unknown: %v", err)
	}
}

func TestReadPacketShortPayload(t *testing.T) {
	if _, err := readPacket(bytes.NewReader(packet(1, []byte{1, 2, 3})[:14])); err != io.ErrUnexpectedEOF {
		t.Fatalf("expected an unexpected EOF, got %v", err)
	}
}

func TestSessionControlMessages(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()
	s := &Session{control: client}
	tests := []struct {
		name string
		send func() error
		size int
		typ  byte
	}{
		{"text", func() error { return s.InjectText("hi") }, 7, typeInjectText},
		{"scroll", func() error { return s.InjectScroll(Position{1, 2, 3, 4}, 1, 1) }, 21, typeInjectScroll},
		{"reset video", s.ResetVideo, 1, typeResetVideo},
	}
	for _, tt := range tests {
		go func() { _ = tt.send() }()
		buf := make([]byte, tt.size)
		if _, err := io.ReadFull(server, buf); err != nil || buf[0] != tt.typ {
			t.Fatalf("%s: %x %v", tt.name, buf, err)
		}
	}
	if err := s.InjectText(strings.Repeat("x", injectTextMaxLength+1)); err == nil {
		t.Fatal("expected an error for too long text")
	}
}

func TestNewLauncherScidIsPositive(t *testing.T) {
	if id := NewLauncher(Deps{}).deps.SessionID(); id <= 0 {
		t.Fatalf("scid = %d", id)
	}
}
