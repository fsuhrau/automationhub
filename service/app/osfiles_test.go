package app

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hubapp "github.com/fsuhrau/automationhub/app"
)

func TestOSFilesPath(t *testing.T) {
	root := t.TempDir()
	files := NewOSFiles(root)
	tests := []struct {
		name, in, want string
	}{
		{"a file", "game.apk", filepath.Join(root, "game.apk")},
		{"empty", "", ""},
		{"the root itself", ".", ""},
		{"leaving the root", "../outside.txt", ""},
		{"a sibling with the root as prefix", "../" + filepath.Base(root) + "x/a", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := files.Path(tt.in); got != tt.want {
				t.Errorf("Path(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestOSFilesSaveMoveReadWriteRemove(t *testing.T) {
	root := filepath.Join(t.TempDir(), "store")
	files := NewOSFiles(root)

	path, err := files.Save("../evil/Game.apk", strings.NewReader("apk"))
	if err != nil || path != filepath.Join(root, "Game.apk") {
		t.Fatalf("save: %q, %v; want the base name inside the root", path, err)
	}
	target := files.Path("moved.apk")
	if err := files.Move(path, target); err != nil {
		t.Fatal(err)
	}
	if data, err := files.Read(target); err != nil || string(data) != "apk" {
		t.Fatalf("read %q, %v", data, err)
	}
	if err := files.Write(target+".icon", []byte("i")); err != nil {
		t.Fatal(err)
	}
	if err := files.Remove(target); err != nil {
		t.Fatal(err)
	}
	if _, err := files.Read(target); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("read after remove: %v", err)
	}
	if err := files.Remove(target); err != nil {
		t.Errorf("removing a missing file: %v", err)
	}
}

func TestOSFilesSaveErrors(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewOSFiles(filepath.Join(blocker, "sub")).Save("a", strings.NewReader("x")); err == nil {
		t.Error("a root that can't be created must fail")
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := NewOSFiles(dir).Save("sub", strings.NewReader("x")); err == nil {
		t.Error("a name that is a folder must fail")
	}
	if _, err := NewOSFiles(dir).Save("a", failingReader{}); err == nil {
		t.Error("a broken upload must fail")
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errBoom }

// writeIPA writes a minimal ipa with an icon.
func writeIPA(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	w := zip.NewWriter(f)
	plist, _ := w.Create("Payload/Game.app/Info.plist")
	_, _ = plist.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict>
<key>CFBundleIdentifier</key><string>com.example.game</string>
<key>CFBundleShortVersionString</key><string>1.2.3</string>
<key>CFBundleVersion</key><string>45</string>
<key>CFBundleDisplayName</key><string>Game</string>
<key>MinimumOSVersion</key><string>15.0</string>
<key>CFBundleIconFiles</key><array><string>Icon</string></array>
</dict></plist>`))
	icon, _ := w.Create("Payload/Game.app/Icon@2x.png")
	_, _ = icon.Write([]byte("\x89PNG\r\n\x1a\nnot crushed"))
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestFileAnalyserReadsAnIPA(t *testing.T) {
	path := filepath.Join(t.TempDir(), "game.ipa")
	writeIPA(t, path)
	var a FileAnalyser

	analysed, err := a.Analyse(path)
	if err != nil || analysed.Identifier != "com.example.game" || analysed.App.Hash == "" || analysed.Info.Icon == "" {
		t.Fatalf("analyse %+v, %v", analysed, err)
	}
	read, err := a.Read(path)
	if err != nil || read.Name != "Game" || read.Platform != "ios" {
		t.Fatalf("read %+v, %v", read, err)
	}
	data, contentType, err := a.Icon(path, analysed.Info.Icon)
	if err != nil || contentType != "image/png" || len(data) == 0 {
		t.Fatalf("icon %d bytes %q, %v", len(data), contentType, err)
	}
}

func TestFileAnalyserErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "none.ipa")
	var a FileAnalyser
	if _, err := a.Analyse(missing); err == nil {
		t.Error("analyse of a missing file must fail")
	}
	if _, err := a.Read(missing); err == nil {
		t.Error("read of a missing file must fail")
	}
	if _, _, err := a.Icon(missing, ""); !errors.Is(err, hubapp.ErrNoIcon) {
		t.Errorf("icon without a path: %v", err)
	}
}
