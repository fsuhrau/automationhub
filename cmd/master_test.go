package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
)

func TestRunMasterReturnsTheListenError(t *testing.T) {
	cfg := sqliteConfig(t)
	cfg.Port = busyPort(t)

	if err := runMaster(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "address already in use") {
		t.Errorf("err = %v", err)
	}
}

func TestRunMasterFailsWithoutDatabase(t *testing.T) {
	err := runMaster(context.Background(), config.Service{Paths: sqliteConfig(t).Paths})

	if err == nil || !strings.Contains(err.Error(), "open database") {
		t.Errorf("err = %v", err)
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
