package api

import (
	"io"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
)

func scopeService(t *testing.T) *Service {
	t.Helper()
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return &Service{db: testkit.DB(t), logger: logrus.NewEntry(logger)}
}

// Devices without a company are visible to every project; the others only
// to the projects of their company.
func TestProjectDevicesAreScopedToTheCompany(t *testing.T) {
	s := scopeService(t)
	shared := testkit.Device(s.db)
	mine := testkit.Device(s.db, func(d *models.Device) { d.CompanyID = 1 })
	testkit.Device(s.db, func(d *models.Device) { d.CompanyID = 2 })

	var got []models.Device
	if err := s.projectDevices(&models.Project{CompanyID: 1}).Order("id").Find(&got).Error; err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 || got[0].ID != shared.ID || got[1].ID != mine.ID {
		t.Fatalf("got %+v", got)
	}
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
