package api

import (
	"io"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/sirupsen/logrus"
)

func scopeService(t *testing.T) *Service {
	t.Helper()
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return &Service{db: testkit.DB(t), logger: logrus.NewEntry(logger)}
}

func TestControlChangeOfAnUnknownDeviceReloadsNothing(t *testing.T) {
	s := scopeService(t)
	r := make(deviceListRecorder, 1)
	events.DeviceListChanged.Register(r)

	s.onDeviceControlChanged("nobody")

	select {
	case p := <-r:
		t.Fatalf("unexpected reload %+v", p)
	case <-time.After(50 * time.Millisecond):
	}
}
