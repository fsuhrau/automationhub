package cmd

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// resetViper isolates a test from the process-wide viper.
func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

// busyPort listens on a free port for the whole test, so a server asked to
// listen there fails at once.
func busyPort(t *testing.T) int32 {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return int32(l.Addr().(*net.TCPAddr).Port)
}

// sqliteConfig is a service config with a database in the test's temp dir.
func sqliteConfig(t *testing.T) config.Service {
	t.Helper()
	return config.Service{
		HostIP:   "127.0.0.1",
		Database: config.Database{SQLite: &config.SQLite{Path: filepath.Join(t.TempDir(), "hub.db")}},
		Paths:    config.Paths{AppStore: t.TempDir(), Bundles: t.TempDir(), TestData: t.TempDir()},
	}
}

// writeConfigFile stores cfg as YAML and returns its path.
func writeConfigFile(t *testing.T, cfg config.Service) string {
	t.Helper()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// execute runs the hub command line and returns what it printed.
func execute(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}
