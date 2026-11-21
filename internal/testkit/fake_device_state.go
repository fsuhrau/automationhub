package testkit

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
)

// Lock locks the device unless a failure is scripted for "Lock".
func (d *FakeDevice) Lock() error {
	if err := d.record("Lock"); err != nil {
		return err
	}
	d.mu.Lock()
	d.locked = true
	d.mu.Unlock()
	return nil
}

// Unlock unlocks the device unless a failure is scripted for "Unlock".
func (d *FakeDevice) Unlock() error {
	if err := d.record("Unlock"); err != nil {
		return err
	}
	d.mu.Lock()
	d.locked = false
	d.mu.Unlock()
	return nil
}

func (d *FakeDevice) IsLocked() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.locked
}

func (d *FakeDevice) SetLogWriter(w device.LogWriter) {
	d.mu.Lock()
	d.writer = w
	d.mu.Unlock()
}

func (d *FakeDevice) GetLogWriter() device.LogWriter {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.writer
}

func (d *FakeDevice) Data(source, data string) { _ = d.record("Data", source, data) }
func (d *FakeDevice) SetKey(key, value string) { _ = d.record("SetKey", key, value) }
func (d *FakeDevice) Passed(result bool)       { _ = d.record("Passed", result) }

//nolint:revive // implements device.Device
func (d *FakeDevice) LogPerformance(checkpoint string, cpu, fps, mem, vertexCount, triangles float64, other string) {
	_ = d.record("LogPerformance", checkpoint, cpu, fps, mem, vertexCount, triangles, other)
}

func (d *FakeDevice) Log(source, format string, params ...interface{}) {
	_ = d.record("Log", source, format, params)
}

func (d *FakeDevice) Error(source, format string, params ...interface{}) {
	_ = d.record("Error", source, format, params)
}

func (d *FakeDevice) Exception(source, format string, params ...interface{}) {
	_ = d.record("Exception", source, format, params)
}

func (d *FakeDevice) AddActionHandler(h action.ActionHandler) {
	d.mu.Lock()
	d.handlers = append(d.handlers, h)
	d.mu.Unlock()
}

// RemoveActionHandler removes the first handler equal to h.
func (d *FakeDevice) RemoveActionHandler(h action.ActionHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, have := range d.handlers {
		if have == h {
			d.handlers = append(d.handlers[:i], d.handlers[i+1:]...)
			return
		}
	}
}

func (d *FakeDevice) ActionHandlers() []action.ActionHandler {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]action.ActionHandler(nil), d.handlers...)
}

// Optional capabilities.

func (d *FakeDevice) Tap(x, y int) error { return d.record("Tap", x, y) }

//nolint:revive // implements device.InputCapable
func (d *FakeDevice) Swipe(fromX, fromY, toX, toY int, duration time.Duration) error {
	return d.record("Swipe", fromX, fromY, toX, toY, duration)
}

func (d *FakeDevice) InputText(text string) error { return d.record("InputText", text) }
func (d *FakeDevice) PressKey(key string) error   { return d.record("PressKey", key) }

// ReadLogs returns the last lines of LogLines.
func (d *FakeDevice) ReadLogs(lines int) ([]string, error) {
	return tail(d.LogLines, lines), d.record("ReadLogs", lines)
}

// ReadLogsSince returns the last lines of LogLines and records since.
func (d *FakeDevice) ReadLogsSince(since string, lines int) ([]string, error) {
	return tail(d.LogLines, lines), d.record("ReadLogsSince", since, lines)
}

// StartVideo records the call and returns a stream of the screen size; it
// fails when "StartVideo" has a scripted error.
func (d *FakeDevice) StartVideo(_ context.Context, opts device.VideoOptions) (device.VideoStream, error) {
	if err := d.record("StartVideo", opts); err != nil {
		return nil, err
	}
	return videoStream(d.Width, d.Height), nil
}

func tail(lines []string, n int) []string {
	if n <= 0 || n >= len(lines) {
		return lines
	}
	return lines[len(lines)-n:]
}
