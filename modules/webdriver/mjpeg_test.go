package webdriver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// wdaChunk is a part as WebDriverAgent's FBMjpegServer writes it.
func wdaChunk(frame []byte) string {
	return fmt.Sprintf("--BoundaryString\r\nContent-type: image/jpg\r\nContent-Length: %d\r\n\r\n%s\r\n\r\n", len(frame), frame)
}

func TestMjpegReaderWDAFormat(t *testing.T) {
	frames := [][]byte{
		{0xFF, 0xD8, 1, 2, '\r', '\n', '-', '-', 3, 0xFF, 0xD9},
		{0xFF, 0xD8, 0xFF, 0xD9},
	}
	r := NewMjpegReader(strings.NewReader(wdaChunk(frames[0]) + wdaChunk(frames[1])))
	for i, want := range frames {
		got, err := r.ReadFrame()
		if err != nil {
			t.Fatalf("frame %d: %v", i, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("frame %d = %x, want %x", i, got, want)
		}
	}
	if _, err := r.ReadFrame(); err != io.EOF {
		t.Fatalf("after the last frame: %v, want EOF", err)
	}
}

func TestMjpegReaderWithoutContentLength(t *testing.T) {
	frame := []byte{0xFF, 0xD8, 7, 0xFF, 0x00, 8, 0xFF, 0xD9}
	body := "--frame\r\nContent-Type: image/jpeg\r\n\r\n" + string(frame) + "\r\n--frame\r\nContent-Type: image/jpeg\r\n\r\n" + string(frame) + "\r\n"
	r := NewMjpegReader(strings.NewReader(body))
	for i := 0; i < 2; i++ {
		got, err := r.ReadFrame()
		if err != nil {
			t.Fatalf("frame %d: %v", i, err)
		}
		if !bytes.Equal(got, frame) {
			t.Fatalf("frame %d = %x, want %x", i, got, frame)
		}
	}
}

func TestMjpegReaderRejectsHugeFrames(t *testing.T) {
	r := NewMjpegReader(strings.NewReader(fmt.Sprintf("--b\r\nContent-Length: %d\r\n\r\n", maxMjpegFrame+1)))
	if _, err := r.ReadFrame(); err == nil {
		t.Fatal("expected an error")
	}
}

func TestMjpegURL(t *testing.T) {
	for address, want := range map[string]string{
		"http://192.168.1.20:8100":  "http://192.168.1.20:9100/",
		"http://[fe80::1]:8100/":    "http://[fe80::1]:9100/",
		"http://localhost:8100/wda": "http://localhost:9100/",
	} {
		got, err := New(address).MjpegURL(DefaultMjpegServerPort)
		if err != nil || got != want {
			t.Errorf("MjpegURL(%q) = %q, %v; want %q", address, got, err, want)
		}
	}
	if _, err := New("not a url").MjpegURL(9100); err == nil {
		t.Error("expected an error for an address without host")
	}
}

func TestOpenMjpegReadsFramesFromAnHTTP10Server(t *testing.T) {
	frame := []byte{0xFF, 0xD8, 42, 0xFF, 0xD9}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Read(make([]byte, 4096))
		// the exact response header of FBMjpegServer
		_, _ = io.WriteString(conn, "HTTP/1.0 200 OK\r\nServer: WDA MJPEG Server\r\nConnection: close\r\nMax-Age: 0\r\nExpires: 0\r\nCache-Control: no-cache, private\r\nPragma: no-cache\r\nContent-Type: multipart/x-mixed-replace; boundary=--BoundaryString\r\n\r\n")
		_, _ = io.WriteString(conn, wdaChunk(frame)+wdaChunk(frame))
	}()
	stream, err := OpenMjpeg(context.Background(), "http://"+ln.Addr().String()+"/")
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	for i := 0; i < 2; i++ {
		got, err := stream.ReadFrame()
		if err != nil || !bytes.Equal(got, frame) {
			t.Fatalf("frame %d = %x, %v", i, got, err)
		}
	}
}

func TestScreenAndUpdateSettings(t *testing.T) {
	var settingsBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/wda/screen":
			_, _ = io.WriteString(w, `{"value":{"statusBarSize":{"width":390,"height":47},"scale":3,"screenSize":{"width":390,"height":844}},"sessionId":"s"}`)
		case "/session/s1/appium/settings":
			b, _ := io.ReadAll(r.Body)
			settingsBody = string(b)
			_, _ = io.WriteString(w, `{"value":{}}`)
		case "/status":
			_, _ = io.WriteString(w, `{"value":{"ready":true}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := New(srv.URL)
	c.sessionId = "s1"
	screen, err := c.Screen()
	if err != nil || screen != (ScreenInfo{Width: 390, Height: 844, Scale: 3}) {
		t.Fatalf("Screen() = %+v, %v", screen, err)
	}
	if err := c.UpdateSettings(map[string]any{"mjpegServerFramerate": 30}); err != nil {
		t.Fatal(err)
	}
	if settingsBody != `{"settings":{"mjpegServerFramerate":30}}` {
		t.Fatalf("settings body = %s", settingsBody)
	}
	if !c.Alive(time.Second) {
		t.Fatal("Alive() = false")
	}
}
