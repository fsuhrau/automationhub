package iossim

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/png" // registers the decoder for the simctl screenshot
)

// StartRecording records the screen to ./<path>.mp4 until StopRecording.
func (d *Device) StartRecording(path string) error {
	d.recMu.Lock()
	defer d.recMu.Unlock()
	p, err := d.deps.Runner.Start(context.Background(), "xcrun", "simctl", "io", d.ID, "recordVideo", fmt.Sprintf("./%s.mp4", path))
	if err != nil {
		return fmt.Errorf("start recording: %w", err)
	}
	d.recording = p
	return nil
}

// StopRecording stops a running recording; without one it does nothing.
func (d *Device) StopRecording() error {
	d.recMu.Lock()
	p := d.recording
	d.recording = nil
	d.recMu.Unlock()
	if p == nil {
		return nil
	}
	if err := p.Kill(); err != nil {
		return fmt.Errorf("stop recording: %w", err)
	}
	return nil
}

// GetScreenshot returns the PNG of the simulator screen and its size.
func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	data, err := d.simctl.run("io", d.ID, "screenshot", "--type=png", "-")
	if err != nil {
		return nil, 0, 0, err
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("decode screenshot: %w", err)
	}
	return data, cfg.Width, cfg.Height, nil
}
