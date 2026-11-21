package device_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestDisplayName(t *testing.T) {
	tests := []struct {
		name string
		dev  *models.Device
		want string
	}{
		{"alias wins", &models.Device{Alias: "Lab", Name: "Pixel", DeviceIdentifier: "dev-1"}, "Lab"},
		{"name without alias", &models.Device{Name: "Pixel", DeviceIdentifier: "dev-1"}, "Pixel"},
		{"identifier as last resort", &models.Device{DeviceIdentifier: "dev-1"}, "dev-1"},
		{"nothing set", &models.Device{}, ""},
		{"nil device", nil, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := device.DisplayName(tc.dev); got != tc.want {
				t.Fatalf("DisplayName = %q, want %q", got, tc.want)
			}
		})
	}
}
