package app

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyseIOSAppFolder(t *testing.T) {
	app := filepath.Join(t.TempDir(), "Game.app")
	writeFile(t, filepath.Join(app, "Info.plist"), infoPlist(t, map[string]interface{}{"CFBundleDisplayName": "Game", "CFBundleVersion": "5"}))
	a := NewAnalyser(app)
	if err := a.AnalyseFile(); err != nil {
		t.Fatal(err)
	}
	p := a.GetParameter()
	if p.Platform != "ios" || p.Name != "Game" || p.Version != "5" || p.Identifier != "com.x" || p.App.Hash != "" {
		t.Errorf("parameter = %+v", p)
	}
}

func TestAnalyseMacAppFolder(t *testing.T) {
	tests := []struct {
		name     string
		plist    map[string]interface{}
		files    []string
		wantExec string
		wantErr  bool
	}{
		{"executable of the plist", map[string]interface{}{}, nil, "Contents/MacOS/Game", false},
		{"first file of MacOS", map[string]interface{}{"CFBundleExecutable": ""}, []string{"a-bin"}, "Contents/MacOS/a-bin", false},
		{"no executable at all", map[string]interface{}{"CFBundleExecutable": ""}, nil, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := filepath.Join(t.TempDir(), "Game.app")
			writeFile(t, filepath.Join(app, "Contents", "Info.plist"), infoPlist(t, tt.plist))
			if err := os.MkdirAll(filepath.Join(app, "Contents", "MacOS", "dir"), 0o755); err != nil {
				t.Fatal(err)
			}
			for _, f := range tt.files {
				writeFile(t, filepath.Join(app, "Contents", "MacOS", f), []byte("bin"))
			}
			a := NewAnalyser(app)
			err := a.AnalyseFile()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v", err)
			}
			p := a.GetParameter()
			if !tt.wantErr && (p.Platform != "macos" || p.Name != "Game.app" || p.App.Executable.Executable != tt.wantExec) {
				t.Errorf("parameter = %+v %+v", p, p.App.Executable)
			}
		})
	}
}

func TestAnalyseAppFolderErrors(t *testing.T) {
	a := NewAnalyser(filepath.Join(t.TempDir(), "None.app"))
	if err := a.AnalyseFile(); err == nil {
		t.Error("a missing app is an error")
	}
}

func TestGetFirstFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := getFirstFile(filepath.Join(dir, "none")); err == nil {
		t.Error("a missing folder is an error")
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := getFirstFile(dir); err == nil {
		t.Error("a folder without files is an error")
	}
	writeFile(t, filepath.Join(dir, "x"), nil)
	if got, err := getFirstFile(dir); err != nil || got != filepath.Join(dir, "x") {
		t.Errorf("file = %s, err = %v", got, err)
	}
}

func TestAnalyseMacArchiveFindsTheExecutable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Game.zip")
	writeZip(t, path, map[string][]byte{
		"Game.app/Contents/Info.plist":  infoPlist(t, map[string]interface{}{"CFBundleExecutable": "", "CFBundleShortVersionString": "2.0"}),
		"Game.app/Contents/MacOS/sub/x": []byte("x"),
		"Game.app/Contents/MacOS/Game":  []byte("bin"),
		"Game.app/Contents/Other":       []byte("o"),
	})
	a := NewAnalyser(path)
	if err := a.AnalyseFile(); err != nil {
		t.Fatal(err)
	}
	p := a.GetParameter()
	if p.App.Executable.Executable != "Contents/MacOS/Game" || p.Version != "2.0" {
		t.Errorf("parameter = %+v %+v", p, p.App.Executable)
	}
	if p.Info.Contents.Files != 4 {
		t.Errorf("contents = %+v, the files of the whole app count", p.Info.Contents)
	}
}

func TestAnalyseArchiveErrors(t *testing.T) {
	dir := t.TempDir()
	notZip := filepath.Join(dir, "a.ipa")
	writeFile(t, notZip, []byte("no zip"))
	noApp := filepath.Join(dir, "b.ipa")
	writeZip(t, noApp, map[string][]byte{"readme": nil})
	for _, path := range []string{notZip, noApp, filepath.Join(dir, "none.ipa")} {
		if err := NewAnalyser(path).AnalyseFile(); err == nil {
			t.Errorf("%s: no error", path)
		}
	}
}

func TestAnalyseUnknownFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.txt")
	writeFile(t, path, []byte("hello"))
	a := NewAnalyser(path)
	if err := a.AnalyseFile(); err != nil {
		t.Fatal(err)
	}
	if p := a.GetParameter(); p.Info.Format != "txt" || p.App.Size != 5 {
		t.Errorf("parameter = %+v info = %+v", p, p.Info)
	}
	if err := NewAnalyser(filepath.Join(t.TempDir(), "none.txt")).AnalyseFile(); err == nil {
		t.Error("a missing file is an error")
	}
}

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "f"), []byte("abc"))
	size, hash, err := hashFile(filepath.Join(dir, "f"))
	if err != nil || size != 3 || hash != "a9993e364706816aba3e25717850c26c9cd0d89d" {
		t.Errorf("size = %d hash = %s err = %v", size, hash, err)
	}
	if _, _, err := hashFile(filepath.Join(dir, "none")); err == nil {
		t.Error("a missing file is an error")
	}
	if _, _, err := hashFile(dir); err == nil {
		t.Error("a folder cannot be read")
	}
}

func TestAnalyseMacArchiveWithoutExecutable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Game.zip")
	writeZip(t, path, map[string][]byte{"Game.app/Contents/Info.plist": infoPlist(t, map[string]interface{}{"CFBundleExecutable": ""})})
	a := NewAnalyser(path)
	if err := a.AnalyseFile(); err != nil {
		t.Fatal(err)
	}
	if got := a.GetParameter().App.Executable.Executable; got != "" {
		t.Errorf("executable = %q", got)
	}
}

func TestAnalyseFileFailsWhenTheFileCannotBeHashed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.txt")
	writeFile(t, path, []byte("hello"))
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	if f, err := os.Open(path); err == nil {
		_ = f.Close()
		t.Skip("the file is readable anyway, running as root")
	}
	if err := NewAnalyser(path).AnalyseFile(); err == nil {
		t.Error("an unreadable file cannot be hashed")
	}
}
