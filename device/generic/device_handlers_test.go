package generic

import (
	"testing"

	"github.com/fsuhrau/automationhub/device/appproto"
)

type namedHandler struct{ name string }

func (*namedHandler) OnActionResponse(interface{}, *appproto.Response) {}

func TestDeviceActionHandlers(t *testing.T) {
	var d Device
	a, b, c := &namedHandler{"a"}, &namedHandler{"b"}, &namedHandler{"c"}
	d.AddActionHandler(a)
	d.AddActionHandler(b)
	d.AddActionHandler(c)

	d.RemoveActionHandler(b)
	d.RemoveActionHandler(b) // unknown handler: no-op

	got := d.ActionHandlers()
	if len(got) != 2 || got[0] != a || got[1] != c {
		t.Errorf("ActionHandlers() = %v", got)
	}
	got[0] = nil // the snapshot is a copy
	if d.ActionHandlers()[0] != a {
		t.Error("ActionHandlers() shares its backing array")
	}
}
