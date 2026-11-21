package device

import (
	"context"

	"github.com/fsuhrau/automationhub/events"
)

// Identities resolves the identifier of a device to its id.
type Identities interface {
	IDByIdentifier(ctx context.Context, identifier string) (uint, error)
}

// Logger is where a notifier reports what it skipped.
type Logger interface {
	Debugf(format string, args ...any)
}

// ListNotifier tells the clients to reload the device list when a device
// got a new live controller, lost it, or was locked or unlocked, so
// "controlled by" and "locked by" show up right away.
type ListNotifier struct {
	ids Identities
	bus *events.Bus[events.DeviceListChangedPayload]
	log Logger
}

// NewListNotifier returns a notifier that looks devices up in ids and
// announces the change on bus.
func NewListNotifier(ids Identities, bus *events.Bus[events.DeviceListChangedPayload], log Logger) *ListNotifier {
	return &ListNotifier{ids: ids, bus: bus, log: log}
}

// DeviceChanged announces a change of the device; an unknown device is
// logged and ignored.
func (n *ListNotifier) DeviceChanged(identifier string) {
	id, err := n.ids.IDByIdentifier(context.Background(), identifier)
	if err != nil {
		n.log.Debugf("change of unknown device %s: %v", identifier, err)
		return
	}
	n.bus.Trigger(events.DeviceListChangedPayload{DeviceID: id})
}
