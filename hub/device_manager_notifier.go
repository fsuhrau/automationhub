package hub

import "github.com/fsuhrau/automationhub/hub/node"

// ChangeNotifier is told whenever the device list may have changed.
type ChangeNotifier interface {
	Notify()
}

// NotifierFactory builds the notifier for a device list snapshot function.
type NotifierFactory func(snapshot func() []*node.DeviceResponse) ChangeNotifier

type noopNotifier struct{}

func (noopNotifier) Notify() {}

// masterNotifierFactory pushes the list to the master; with no master the
// node notifier is nil and Notify on it does nothing.
func (dm *DeviceManager) masterNotifierFactory(snapshot func() []*node.DeviceResponse) ChangeNotifier {
	return node.NewDevicesNotifier(dm.masterUrl, dm.nodeIdentifier, dm.authToken, snapshot)
}
