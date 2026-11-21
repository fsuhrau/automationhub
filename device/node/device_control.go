package node

import (
	"context"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/google/uuid"
)

func (d *NodeDevice) control(req device.ControlRequest) (device.ControlResponse, error) {
	req.DeviceID = d.ID
	return d.nodeManager.Control(d.nodeId, req)
}

// Tap implements device.InputCapable by forwarding to the node.
func (d *NodeDevice) Tap(x, y int) error {
	_, err := d.control(device.ControlRequest{Op: device.ControlTap, X: x, Y: y})
	return err
}

// Swipe implements device.InputCapable by forwarding to the node.
//
//nolint:revive // argument-limit: fixed by device.InputCapable
func (d *NodeDevice) Swipe(fromX, fromY, toX, toY int, duration time.Duration) error {
	_, err := d.control(device.ControlRequest{Op: device.ControlSwipe, X: fromX, Y: fromY, ToX: toX, ToY: toY, DurationMs: int(duration.Milliseconds())})
	return err
}

// InputText implements device.InputCapable by forwarding to the node.
func (d *NodeDevice) InputText(text string) error {
	_, err := d.control(device.ControlRequest{Op: device.ControlText, Text: text})
	return err
}

// PressKey implements device.InputCapable by forwarding to the node.
func (d *NodeDevice) PressKey(key string) error {
	_, err := d.control(device.ControlRequest{Op: device.ControlKey, Key: key})
	return err
}

// ReadLogs implements device.LogReader by forwarding to the node.
func (d *NodeDevice) ReadLogs(lines int) ([]string, error) {
	resp, err := d.control(device.ControlRequest{Op: device.ControlReadLogs, Lines: lines})
	return resp.Lines, err
}

// ReadLogsSince implements device.LogFollower by forwarding to the node.
func (d *NodeDevice) ReadLogsSince(since string, lines int) ([]string, error) {
	resp, err := d.control(device.ControlRequest{Op: device.ControlReadLogs, Lines: lines, Since: since})
	return resp.Lines, err
}

// videoConnectTimeout bounds how long the master waits for the node to
// connect the stream back. iOS devices may start WebDriverAgent first
// (iosdevice.XcodebuildStartTimeout).
const videoConnectTimeout = 150 * time.Second

// StartVideo implements device.VideoStreamer. The node starts the stream and
// dials back to the master, which pairs the socket through the rendezvous.
func (d *NodeDevice) StartVideo(ctx context.Context, opts device.VideoOptions) (device.VideoStream, error) {
	streamID := uuid.NewString()
	ch, done := d.rendezvous.Expect(streamID)
	defer done()
	request := device.VideoStreamRequest{DeviceID: d.ID, StreamID: streamID, Options: opts}
	if err := d.nodeManager.StartVideoStream(d.nodeId, request); err != nil {
		return nil, err
	}
	stream, err := videostream.Await(ctx, ch, videoConnectTimeout)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		_ = stream.Close()
	}()
	return stream, nil
}

func (d *NodeDevice) IsAppConnected() bool {
	return d.nodeManager.IsConnected(d.nodeId, d.ID)
}

func (d *NodeDevice) ConnectionTimeout() time.Duration {
	return d.nodeManager.ConnectionTimeout(d.nodeId, d.ID)
}

// Send forwards the action to the node. The node reports no delivery result.
func (d *NodeDevice) Send(script []byte) error {
	d.nodeManager.SendAction(d.nodeId, d.ID, script)
	return nil
}
