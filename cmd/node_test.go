package cmd

import (
	"context"
	"strings"
	"testing"
)

func TestRunNodeNeedsAMasterURL(t *testing.T) {
	if err := runNode(context.Background(), sqliteConfig(t)); err == nil || !strings.Contains(err.Error(), "master url") {
		t.Errorf("err = %v", err)
	}
}

func TestNodeCommandRunsWithTheConfigFile(t *testing.T) {
	resetViper(t)

	_, err := execute(t, "--config", writeConfigFile(t, sqliteConfig(t)), "node")

	if err == nil || !strings.Contains(err.Error(), "master url") {
		t.Errorf("err = %v", err)
	}
}

func TestRunNodeReturnsTheListenError(t *testing.T) {
	cfg := sqliteConfig(t)
	cfg.MasterURL = "127.0.0.1:1"
	cfg.Port = busyPort(t)

	if err := runNode(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "address already in use") {
		t.Errorf("err = %v", err)
	}
}
