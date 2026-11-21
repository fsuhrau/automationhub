package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/spf13/viper"
)

func TestLoadServiceConfigReadsViper(t *testing.T) {
	resetViper(t)
	viper.Set("identifier", "hub-1")
	viper.Set("log", "warn")

	cfg, err := loadServiceConfig()

	if err != nil || cfg.Identifier != "hub-1" || cfg.LogLevel != "warn" {
		t.Errorf("cfg = %+v, err = %v", cfg, err)
	}
}

func TestLoadServiceConfigKeepsALevelFromTheFile(t *testing.T) {
	resetViper(t)
	t.Setenv("LOG", "error")
	viper.AutomaticEnv()
	viper.Set("log", "debug")

	cfg, err := loadServiceConfig()

	if err != nil || cfg.LogLevel != "debug" {
		t.Errorf("level = %q, err = %v", cfg.LogLevel, err)
	}
}

func TestLoadServiceConfigFailsOnAWrongType(t *testing.T) {
	resetViper(t)
	viper.Set("port", "not-a-number")

	if _, err := loadServiceConfig(); err == nil || !strings.Contains(err.Error(), "read config") {
		t.Errorf("err = %v", err)
	}
}

func TestLoadServiceConfigFillsTheFoldersBelowHome(t *testing.T) {
	resetViper(t)
	t.Setenv("HOME", "/home/hub")
	viper.Set("paths.test_data", "/data")

	cfg, err := loadServiceConfig()

	want := config.Paths{
		AppStore: filepath.Join("/home/hub", ".automationhub", "store"),
		Bundles:  filepath.Join("/home/hub", ".automationhub", "bundles"),
		TestData: "/data",
	}
	if err != nil || cfg.Paths != want {
		t.Errorf("paths = %+v, err = %v", cfg.Paths, err)
	}
}
