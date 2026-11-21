package cmd

import (
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/gin-gonic/gin"
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

func TestGetHostIP(t *testing.T) {
	tests := map[string]struct {
		hostIP string
		want   string
	}{
		"configured": {"10.1.2.3", "10.1.2.3"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := getHostIP(config.Service{HostIP: tc.hostIP}); got.String() != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGetHostIPFallsBackToTheOutboundIP(t *testing.T) {
	for name, hostIP := range map[string]string{"empty": "", "invalid": "nope"} {
		t.Run(name, func(t *testing.T) {
			if got := getHostIP(config.Service{HostIP: hostIP}); net.IP(got) == nil {
				t.Error("no ip")
			}
		})
	}
}

type fakeEndpoint struct{}

func (fakeEndpoint) RegisterRoutes(*gin.Engine, *gin.RouterGroup) error { return nil }

func TestAddEndpointsRegistersInOrder(t *testing.T) {
	adder := &fakeEndpointAdder{}

	if err := addEndpoints(adder, fakeEndpoint{}, fakeEndpoint{}); err != nil {
		t.Fatal(err)
	}

	if len(adder.added) != 2 {
		t.Errorf("added = %d", len(adder.added))
	}
}

func TestAddEndpointsStopsAtTheFirstFailure(t *testing.T) {
	boom := errors.New("boom")
	adder := &fakeEndpointAdder{fail: boom}

	err := addEndpoints(adder, fakeEndpoint{}, fakeEndpoint{})

	if !errors.Is(err, boom) || len(adder.added) != 0 {
		t.Errorf("err = %v, added = %d", err, len(adder.added))
	}
}
