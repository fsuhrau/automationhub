package device_test

import (
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

var _ device.Clock = (*testkit.FakeClock)(nil)

func TestRealClock(t *testing.T) {
	var c device.Clock = device.RealClock{}
	before := time.Now()
	if now := c.Now(); now.Before(before) {
		t.Errorf("Now() = %v, before %v", now, before)
	}
	select {
	case <-c.After(time.Millisecond):
	case <-time.After(5 * time.Second):
		t.Fatal("After did not fire")
	}
}
