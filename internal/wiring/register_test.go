package wiring

import (
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/endpoints"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
)

type fakeEndpoint struct{}

func (fakeEndpoint) RegisterRoutes(*gin.Engine, *gin.RouterGroup) error { return nil }

type fakeAdder struct {
	added []endpoints.ServiceEndpoint
	fail  error
}

func (f *fakeAdder) AddEndpoint(ep endpoints.ServiceEndpoint) error {
	if f.fail != nil {
		return f.fail
	}
	f.added = append(f.added, ep)
	return nil
}

func TestAddEndpointsRegistersInOrder(t *testing.T) {
	adder := &fakeAdder{}

	if err := addEndpoints(adder, fakeEndpoint{}, fakeEndpoint{}); err != nil {
		t.Fatal(err)
	}

	if len(adder.added) != 2 {
		t.Errorf("added = %d", len(adder.added))
	}
}

func TestAddEndpointsStopsAtTheFirstFailure(t *testing.T) {
	boom := errors.New("boom")
	adder := &fakeAdder{fail: boom}

	err := addEndpoints(adder, fakeEndpoint{}, fakeEndpoint{})

	if !errors.Is(err, boom) || len(adder.added) != 0 {
		t.Errorf("err = %v, added = %d", err, len(adder.added))
	}
}

func TestManagerNamesAreSorted(t *testing.T) {
	cfg := config.Service{DeviceManager: map[string]config.Manager{"b": {}, "a": {}}}

	got := managerNames(cfg)

	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v", got)
	}
	if len(managerNames(config.Service{})) != 0 {
		t.Error("names without managers")
	}
}

func TestAddHandlersAddsEveryHandler(t *testing.T) {
	devices := &testkit.FakeDeviceManager{}

	addHandlers(devices, []device.Handler{nil, nil})

	if len(devices.CallsTo("AddHandler")) != 2 {
		t.Errorf("calls %v", devices.Calls())
	}
}
