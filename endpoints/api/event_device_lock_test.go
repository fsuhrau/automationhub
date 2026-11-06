package api

import (
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type deviceListRecorder chan events.DeviceListChangedPayload

func (r deviceListRecorder) Handle(p events.DeviceListChangedPayload) { r <- p }

// A lock change of a device (test run, unlock button, MCP) must make the UI
// reload the device list, which carries isLocked and lockedBy.
func TestDeviceLockChangeReloadsDeviceList(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Device{}); err != nil {
		t.Fatal(err)
	}
	db.Create(&models.Device{DeviceIdentifier: "other"})
	phone := models.Device{DeviceIdentifier: "phone"}
	db.Create(&phone)

	r := make(deviceListRecorder, 8)
	events.DeviceListChanged.Register(r)
	s := &Service{db: db, logger: logrus.NewEntry(logrus.New())}
	RegisterEventDeviceLockListener(s)

	for _, locked := range []bool{true, false} {
		events.DeviceLockChanged.Trigger(events.DeviceLockChangedPayload{DeviceIdentifier: "phone", Locked: locked})
		deadline := time.After(time.Second)
	wait:
		for {
			select {
			case p := <-r:
				if p.DeviceID == phone.ID {
					break wait
				}
			case <-deadline:
				t.Fatalf("no device_list event for device %d after lock=%v", phone.ID, locked)
			}
		}
	}
}
