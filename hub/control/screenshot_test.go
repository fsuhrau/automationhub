package control

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

func TestScreenshotOfADeviceWithAScreenshotter(t *testing.T) {
	dev := testkit.NewFakeDevice("d1")
	dev.PNG, dev.Width, dev.Height = []byte("png"), 3, 4
	data, w, h, err := New(nil).Screenshot(dev)
	if err != nil || string(data) != "png" || w != 3 || h != 4 {
		t.Errorf("Screenshot = %q %d %d %v", data, w, h, err)
	}
}

func TestScreenshotOfADeviceWithoutScreenshotterIsUnsupported(t *testing.T) {
	_, _, _, err := New(nil).Screenshot(testkit.NewCoreDevice("d1"))
	if !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("err = %v, want one wrapping device.ErrUnsupported", err)
	}
}
