package scrcpy

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Version of the embedded server. The server refuses to start when the
// client passes a different version.
const Version = "3.3.4"

//go:embed scrcpy-server-v3.3.4
var serverJar []byte

// jarFile writes the embedded server to a temp file once per Launcher.
type jarFile struct {
	once sync.Once
	path string
	err  error
}

func (j *jarFile) local() (string, error) {
	j.once.Do(func() { j.path, j.err = writeJar(serverJar) })
	return j.path, j.err
}

func writeJar(data []byte) (string, error) {
	path := filepath.Join(os.TempDir(), "automationhub-scrcpy-server-v"+Version+".jar")
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, data) {
		return path, nil
	}
	if err := replaceFile(path, data); err != nil {
		return "", fmt.Errorf("write scrcpy server: %w", err)
	}
	return path, nil
}

// replaceFile writes next to path first, so a concurrent reader never sees a
// partial jar.
func replaceFile(path string, data []byte) error {
	tmp := fmt.Sprintf("%s.%d.tmp", path, os.Getpid())
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
