package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
)

func TestBuildMasterWiresTheServices(t *testing.T) {
	cfg := sqliteConfig(t)
	cfg.Hooks = []config.Hook{{Provider: "slack"}}

	m, err := buildMaster(cfg)

	if err != nil || m.server == nil || m.nodes == nil || m.sessions == nil {
		t.Errorf("master = %+v, err = %v", m, err)
	}
}

func TestBuildMasterFailsWithoutDatabase(t *testing.T) {
	if _, err := buildMaster(config.Service{}); err == nil || !strings.Contains(err.Error(), "open database") {
		t.Errorf("err = %v", err)
	}
}

func TestRunMasterReturnsTheListenError(t *testing.T) {
	cfg := sqliteConfig(t)
	cfg.Port = busyPort(t)

	if err := runMaster(cfg); err == nil || !strings.Contains(err.Error(), "address already in use") {
		t.Errorf("err = %v", err)
	}
}

func TestRunMasterFailsWithoutDatabase(t *testing.T) {
	if err := runMaster(config.Service{}); err == nil {
		t.Error("no error")
	}
}

func TestMasterCommandRunsWithTheConfigFile(t *testing.T) {
	resetViper(t)
	cfg := sqliteConfig(t)
	cfg.Port = busyPort(t)

	_, err := execute(t, "--config", writeConfigFile(t, cfg), "master")

	if err == nil || !strings.Contains(err.Error(), "address already in use") {
		t.Errorf("err = %v", err)
	}
}

func TestMasterAndNodeCommandsFailOnABrokenConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.yaml")
	if err := os.WriteFile(path, []byte("port: not-a-number\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"master", "node"} {
		t.Run(command, func(t *testing.T) {
			resetViper(t)
			if _, err := execute(t, "--config", path, command); err == nil || !strings.Contains(err.Error(), "read config") {
				t.Errorf("err = %v", err)
			}
		})
	}
}
