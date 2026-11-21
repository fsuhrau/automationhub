package cmd

import (
	"sort"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
)

func TestManagerNames(t *testing.T) {
	cfg := config.Service{DeviceManager: map[string]config.Manager{"b": {}, "a": {}}}

	got := managerNames(cfg)

	sort.Strings(got)
	if strings.Join(got, ",") != "a,b" {
		t.Errorf("got %v", got)
	}
	if len(managerNames(config.Service{})) != 0 {
		t.Error("names without managers")
	}
}

func TestBuildNodeWiresTheServices(t *testing.T) {
	n, err := buildNode(sqliteConfig(t))

	if err != nil || n.server == nil || n.handler == nil || n.handler.ConnectFunction == nil {
		t.Errorf("node = %+v, err = %v", n, err)
	}
}

func TestBuildNodeFailsWithoutDatabase(t *testing.T) {
	if _, err := buildNode(config.Service{}); err == nil || !strings.Contains(err.Error(), "open database") {
		t.Errorf("err = %v", err)
	}
}

func TestRunNodeNeedsAMasterURL(t *testing.T) {
	if err := runNode(sqliteConfig(t)); err == nil || !strings.Contains(err.Error(), "master url") {
		t.Errorf("err = %v", err)
	}
}

func TestRunNodeFailsWithoutDatabase(t *testing.T) {
	if err := runNode(config.Service{}); err == nil {
		t.Error("no error")
	}
}

func TestNodeCommandRunsWithTheConfigFile(t *testing.T) {
	resetViper(t)

	_, err := execute(t, "--config", writeConfigFile(t, sqliteConfig(t)), "node")

	if err == nil || !strings.Contains(err.Error(), "master url") {
		t.Errorf("err = %v", err)
	}
}
