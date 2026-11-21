package session

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const sessionDir = "logs/1700000000/sid"

func TestNewStorageCreatesDirectoryAndLog(t *testing.T) {
	fs := newMemFS()
	s := NewStorage(deps(fs), "logs", "sid")
	if s.GetPath() != sessionDir {
		t.Errorf("path %q, want %q", s.GetPath(), sessionDir)
	}
	if len(fs.dirs) != 1 || fs.dirs[0] != sessionDir {
		t.Errorf("dirs %v", fs.dirs)
	}
	if _, ok := fs.files[sessionDir+"/remote_device.log"]; !ok {
		t.Errorf("device log not opened: %v", fs.files)
	}
}

func TestStoreFilesAreNamedByTime(t *testing.T) {
	cases := []struct {
		name  string
		store func(*Storage) (string, error)
		want  string
	}{
		{"scene graph", func(s *Storage) (string, error) { return s.StoreSceneGraph([]byte("<a/>")) }, sessionDir + "/1700000000_sceengraph.xml"},
		{"image", func(s *Storage) (string, error) { return s.StoreImage([]byte("<a/>")) }, sessionDir + "/1700000000_screenshot.png"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := newMemFS()
			path, err := c.store(NewStorage(deps(fs), "logs", "sid"))
			if err != nil || path != c.want {
				t.Fatalf("path %q err %v, want %q", path, err, c.want)
			}
			if fs.content(path) != "<a/>" || !fs.files[path].closed {
				t.Errorf("file %q content %q closed %v", path, fs.content(path), fs.files[path].closed)
			}
		})
	}
}

func TestStoreReportsFileSystemErrors(t *testing.T) {
	cases := []struct {
		name string
		fail func(*memFS)
		want string
	}{
		{"mkdir", func(m *memFS) { m.mkdirErr = errDisk }, "create " + sessionDir},
		{"create", func(m *memFS) {
			m.openErr = func(n string) error {
				if strings.HasSuffix(n, ".png") {
					return errDisk
				}
				return nil
			}
		}, "create " + sessionDir + "/1700000000_screenshot.png"},
		{"write", func(m *memFS) { m.writeErr = errDisk }, "write " + sessionDir + "/1700000000_screenshot.png"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs := newMemFS()
			s := NewStorage(deps(fs), "logs", "sid")
			c.fail(fs)
			path, err := s.StoreImage([]byte("x"))
			if path != "" || !errors.Is(err, errDisk) || !strings.HasPrefix(err.Error(), c.want) {
				t.Errorf("path %q err %v, want prefix %q", path, err, c.want)
			}
		})
	}
}

func TestRemoteDeviceLogAppendsTimestampedLines(t *testing.T) {
	fs := newMemFS()
	clock := newClock()
	s := NewStorage(Deps{Clock: clock, FS: fs}, "logs", "sid")
	if err := s.RemoteDeviceLog("hello"); err != nil {
		t.Fatal(err)
	}
	clock.Advance(time.Second)
	if err := s.RemoteDeviceLog("again"); err != nil {
		t.Fatal(err)
	}
	want := "1700000000000000000 hello\n1700000001000000000 again\n"
	if got := fs.content(sessionDir + "/remote_device.log"); got != want {
		t.Errorf("log %q, want %q", got, want)
	}
	if err := s.Close(); err != nil || fs.closeCalls != 1 {
		t.Errorf("close err %v, calls %d", err, fs.closeCalls)
	}
}

func TestRemoteDeviceLogReportsWriteError(t *testing.T) {
	fs := newMemFS()
	s := NewStorage(deps(fs), "logs", "sid")
	fs.writeErr = errDisk
	if err := s.RemoteDeviceLog("x"); !errors.Is(err, errDisk) {
		t.Errorf("err %v, want %v", err, errDisk)
	}
}

func TestStorageWithoutDeviceLogDropsLines(t *testing.T) {
	fs := newMemFS()
	fs.openErr = func(string) error { return errDisk }
	s := NewStorage(deps(fs), "logs", "sid")
	if err := s.RemoteDeviceLog("x"); err != nil {
		t.Errorf("log err %v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("close err %v", err)
	}
}

func TestOSFSAndSystemClock(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b")
	s := NewStorage(SystemDeps(), dir, "sid")
	path, err := s.StoreImage([]byte("png"))
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "png" {
		t.Errorf("file %q err %v", data, err)
	}
	if err := s.RemoteDeviceLog("x"); err != nil || s.Close() != nil {
		t.Errorf("device log failed: %v", err)
	}
	if time.Since(SystemClock{}.Now()) > time.Minute {
		t.Error("system clock is off")
	}
	select {
	case <-SystemClock{}.After(time.Millisecond):
	case <-time.After(5 * time.Second):
		t.Error("After never fired")
	}
	if _, err := (OSFS{}).OpenFile(filepath.Join(dir, "missing", "f"), os.O_RDWR, 0); err == nil {
		t.Error("opening a missing file must fail")
	}
}
