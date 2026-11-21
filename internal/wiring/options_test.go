package wiring

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/sirupsen/logrus"
)

func TestResolveAppliesTheOptionsAndCreatesTheFolders(t *testing.T) {
	root := t.TempDir()
	cfg := config.Service{Paths: config.Paths{
		AppStore: filepath.Join(root, "a"), Bundles: filepath.Join(root, "b"), TestData: filepath.Join(root, "c", "d"),
	}}
	db := testkit.DB(t)
	logger := logrus.New()

	o, err := resolve(cfg, []Option{WithDB(db), WithLogger(logger)})

	if err != nil || o.db != db || o.logger != logger {
		t.Fatalf("options %+v, err %v", o, err)
	}
	for _, dir := range cfg.Paths.All() {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Errorf("folder %s: %v", dir, err)
		}
	}
}

func TestResolveDefaultsTheLogger(t *testing.T) {
	o, err := resolve(config.Service{Paths: config.Paths{AppStore: t.TempDir(), Bundles: t.TempDir(), TestData: t.TempDir()}}, nil)

	if err != nil || o.logger == nil || o.db != nil {
		t.Fatalf("options %+v, err %v", o, err)
	}
}

func TestResolveFailsWhenAFolderCannotBeCreated(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Service{Paths: config.Paths{AppStore: filepath.Join(file, "sub")}}

	if _, err := resolve(cfg, nil); err == nil {
		t.Fatal("no error")
	}
}

func TestOpenDBKeepsAnInjectedDatabase(t *testing.T) {
	db := testkit.DB(t)
	o := options{db: db}

	if err := o.openDB(config.Database{}); err != nil || o.db != db {
		t.Fatalf("db %v, err %v", o.db, err)
	}
}
