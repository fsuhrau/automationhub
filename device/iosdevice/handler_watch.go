package iosdevice

import (
	"context"

	"github.com/sirupsen/logrus"
)

// Start subscribes to usbmuxd attach/detach events, so iPhones show up and
// disappear right away instead of with the next poll.
func (m *Handler) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.stopWatch = cancel
	m.mu.Unlock()
	ev := usbEvents{open: m.deps.openListener, onChange: func() { m.onUsbChange(ctx) }, connected: m.tracking.Store}
	go watchUsbmux(ctx, ev)
	return nil
}

// onUsbChange refreshes right away and again once a freshly attached device
// is ready.
func (m *Handler) onUsbChange(ctx context.Context) {
	m.refreshAndNotify()
	go func() {
		select {
		case <-ctx.Done():
		case <-m.deps.Clock.After(settleDelay):
			m.refreshAndNotify()
		}
	}()
}

func (m *Handler) refreshAndNotify() {
	if err := m.RefreshDevices(true); err != nil {
		logrus.Debugf("ios device refresh: %v", err)
	}
	m.changed()
}

func (m *Handler) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopWatch != nil {
		m.stopWatch()
		m.stopWatch = nil
	}
	return nil
}
