package node

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

func TestNewRPCNodeOpensTheBundleFolderOfTheConfig(t *testing.T) {
	cfg := config.Service{Paths: config.Paths{Bundles: t.TempDir()}}

	n, err := NewRPCNode(cfg, &testkit.FakeDeviceManager{}, nil)

	if err != nil || n == nil || n.abm == nil {
		t.Fatalf("node %+v, err %v", n, err)
	}
}

func TestNewRPCNodeFailsWithoutABundleFolder(t *testing.T) {
	cfg := config.Service{Paths: config.Paths{Bundles: filepath.Join(t.TempDir(), "missing")}}

	n, err := NewRPCNode(cfg, &testkit.FakeDeviceManager{}, nil)

	if n != nil || err == nil || !strings.Contains(err.Error(), "bundle store") {
		t.Fatalf("node %+v, err %v", n, err)
	}
}
