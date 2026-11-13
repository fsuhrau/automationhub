package mcp

import (
	"bytes"
	"image"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

// screenshotSetup has two projects, each with a run whose protocol took
// screenshots, and an unrelated file in the data folder.
func screenshotSetup(t *testing.T) (*httptest.Server, uint, uint) {
	dir := t.TempDir()
	old := apps.TestDataPath
	apps.TestDataPath = dir
	t.Cleanup(func() { apps.TestDataPath = old })

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Project{}, &models.App{}, &models.Test{}, &models.TestRun{}, &models.TestProtocol{}, &models.ProtocolKey{}, &models.ProtocolEntry{}); err != nil {
		t.Fatal(err)
	}

	protocol := func(projectID uint) uint {
		a := models.App{ProjectID: projectID}
		db.Create(&a)
		test := models.Test{AppID: a.ID}
		db.Create(&test)
		run := models.TestRun{TestID: test.ID}
		db.Create(&run)
		p := models.TestProtocol{TestRunID: run.ID, TestName: "Suite/A"}
		db.Create(&p)
		return p.ID
	}
	own, foreign := protocol(1), protocol(2)

	write := func(name string, w, h int) {
		if err := os.WriteFile(filepath.Join(dir, name), testPNG(t, w, h), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("first.png", 400, 800)
	write("last.png", 2000, 1000)
	write("foreign.png", 10, 10)
	write("unreferenced.png", 10, 10)
	db.Create(&models.ProtocolEntry{TestProtocolID: own, Source: "screen", Data: "first.png", Message: "login", Runtime: 1.5})
	db.Create(&models.ProtocolEntry{TestProtocolID: own, Source: "log", Message: "no data"})
	db.Create(&models.ProtocolEntry{TestProtocolID: own, Source: "screen", Data: "last.png", Runtime: 9})
	db.Create(&models.ProtocolEntry{TestProtocolID: own, Source: "data", Data: "trace.json"})
	db.Create(&models.ProtocolEntry{TestProtocolID: foreign, Source: "screen", Data: "foreign.png"})
	db.Create(&models.ProtocolKey{TestProtocolID: own, Key: "player_id", Value: "1234"})

	h := NewHandler(Deps{DB: db, Backend: fakeBackend{}, Control: hubcontrol.New(nil), Runs: runs.NewTracker(nil)})
	project := &models.Project{Identifier: "p1"}
	project.ID = 1
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r, project, r.Header.Get("X-Caller"))
	}))
	t.Cleanup(srv.Close)
	return srv, own, foreign
}

func imageWidth(t *testing.T, data []byte) int {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return cfg.Width
}

func TestProtocolScreenshot(t *testing.T) {
	srv, own, foreign := screenshotSetup(t)
	cs := connect(t, srv, "alice")
	defer cs.Close()

	image := func(args map[string]any) (int, string) {
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

	// by file, with the step it belongs to
	if w, note := image(map[string]any{"protocolId": own, "file": "first.png"}); w != 400 || !strings.Contains(note, "login") {
		t.Fatalf("first.png: width %d, note %q", w, note)
	}
	// without a file: the last screenshot, downscaled
	if w, note := image(map[string]any{"protocolId": own}); w != defaultScreenshotWidth || !strings.Contains(note, "last.png") {
		t.Fatalf("last: width %d, note %q", w, note)
	}
	if w, _ := image(map[string]any{"protocolId": own, "maxWidth": 500}); w != 500 {
		t.Fatalf("maxWidth: width %d", w)
	}

	// get_protocol lists them, and the protocol's keys
	if s := text(call(t, cs, "get_protocol", map[string]any{"protocolId": own})); !strings.Contains(s, "first.png") || !strings.Contains(s, "last.png") || strings.Contains(s, "trace.json") || !strings.Contains(s, `"player_id": "1234"`) {
		t.Fatalf("get_protocol screenshots: %s", s)
	}

	// nothing outside the protocol's own screenshots
	for _, args := range []map[string]any{
		{"protocolId": own, "file": "unreferenced.png"},
		{"protocolId": own, "file": "foreign.png"},
		{"protocolId": own, "file": "../" + filepath.Base(t.TempDir()) + "/first.png"},
		{"protocolId": own, "file": "trace.json"},
		{"protocolId": foreign},
	} {
		if res := call(t, cs, "get_protocol_screenshot", args); !res.IsError {
			t.Errorf("%v: expected an error", args)
		}
	}
}
