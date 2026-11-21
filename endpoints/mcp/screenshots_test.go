package mcp

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// shotsProtocol stores a protocol of the first run that took two screenshots
// and wrote a data file, next to the unrelated files of the data folder.
func shotsProtocol(t *testing.T, w *goldenWorld) uint {
	t.Helper()
	p := testkit.Protocol(w.db, func(p *models.TestProtocol) { p.TestRunID = 1 })
	for _, e := range []models.ProtocolEntry{
		{Source: "screen", Data: "shot-first.png", Message: "login", Runtime: 1.5},
		{Source: "log", Message: "no data"},
		{Source: "screen", Data: "shot-last.png", Runtime: 9},
		{Source: "data", Data: "trace.json"},
	} {
		e.TestProtocolID = p.ID
		w.db.Create(&e)
	}
	w.db.Create(&models.ProtocolKey{TestProtocolID: p.ID, Key: "player_id", Value: "1234"})
	for name, size := range map[string][2]int{"shot-first.png": {400, 800}, "shot-last.png": {2000, 1000}, "unreferenced.png": {10, 10}} {
		if err := os.WriteFile(filepath.Join(w.dataDir, name), testPNG(t, size[0], size[1]), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return p.ID
}

func imageWidth(t *testing.T, data []byte) int {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return cfg.Width
}

// screenshotOf calls get_protocol_screenshot and returns the width of the
// image and the note.
func screenshotOf(t *testing.T, cs *mcp.ClientSession, args map[string]any) (int, string) {
	t.Helper()
	res := call(t, cs, "get_protocol_screenshot", args)
	if res.IsError {
		t.Fatalf("%v: %s", args, text(res))
	}
	for _, c := range res.Content {
		if ic, ok := c.(*mcp.ImageContent); ok && ic.MIMEType == "image/jpeg" {
			return imageWidth(t, ic.Data), text(res)
		}
	}
	t.Fatalf("%v: no jpeg", args)
	return 0, ""
}

func TestProtocolScreenshot(t *testing.T) {
	srv, w := setup(t)
	own := shotsProtocol(t, w)
	cs := connect(t, srv, "alice")

	// by file, with the step it belongs to
	if w, note := screenshotOf(t, cs, a("protocolId", own, "file", "shot-first.png")); w != 400 || !strings.Contains(note, "login") {
		t.Fatalf("first: width %d, note %q", w, note)
	}
	// without a file: the last screenshot, downscaled
	if w, note := screenshotOf(t, cs, a("protocolId", own)); w != defaultScreenshotWidth || !strings.Contains(note, "shot-last.png") {
		t.Fatalf("last: width %d, note %q", w, note)
	}
	if w, _ := screenshotOf(t, cs, a("protocolId", own, "maxWidth", 500)); w != 500 {
		t.Fatalf("maxWidth: width %d", w)
	}
	// get_protocol lists them, and the protocol's keys
	if s := text(call(t, cs, "get_protocol", a("protocolId", own))); !strings.Contains(s, "shot-first.png") || !strings.Contains(s, "shot-last.png") || strings.Contains(s, "trace.json") || !strings.Contains(s, `"player_id": "1234"`) {
		t.Fatalf("get_protocol screenshots: %s", s)
	}
}

func TestProtocolScreenshotServesNothingOutsideTheProtocolsOwn(t *testing.T) {
	srv, w := setup(t)
	own := shotsProtocol(t, w)
	cs := connect(t, srv, "alice")

	for _, args := range []map[string]any{
		{"protocolId": own, "file": "unreferenced.png"},
		{"protocolId": own, "file": "foreign.png"},
		{"protocolId": own, "file": "../" + filepath.Base(t.TempDir()) + "/first.png"},
		{"protocolId": own, "file": "trace.json"},
		{"protocolId": 4},
	} {
		if res := call(t, cs, "get_protocol_screenshot", args); !res.IsError {
			t.Errorf("%v: expected an error", args)
		}
	}
}

func TestPickScreenshot(t *testing.T) {
	entries := []models.ProtocolEntry{
		{Data: "a.png", Message: "old"}, {Data: "trace.json"}, {Data: "a.png", Message: "new"}, {Data: "B.JPG"}, {Data: "c.txt"},
	}
	tests := []struct {
		name, file, want string
		message          string
		fails            bool
	}{
		{"newest screenshot", "", "B.JPG", "", false},
		{"newest entry of the file", "a.png", "a.png", "new", false},
		{"any entry of the file, even a data file", "trace.json", "trace.json", "", false},
		{"file without an entry", "z.png", "", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := pickScreenshot(entries, ProtocolScreenshotArgs{ProtocolID: 7, File: tc.file})

			if tc.fails != (err != nil) || (err == nil && (got.Data != tc.want || got.Message != tc.message)) {
				t.Fatalf("got %+v, err %v", got, err)
			}
		})
	}
	if _, err := pickScreenshot(nil, ProtocolScreenshotArgs{ProtocolID: 7}); err == nil || !strings.Contains(err.Error(), "protocol 7 has no screenshots") {
		t.Errorf("no entries: %v", err)
	}
}
