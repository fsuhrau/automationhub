// Package adb runs adb commands for one device through a device.CommandRunner.
package adb

import (
	"context"
	"io"

	"github.com/fsuhrau/automationhub/device"
)

// Binary is the adb executable.
const Binary = "adb"

// Client runs adb commands against the device with the serial.
type Client struct {
	Runner device.CommandRunner
	Serial string
}

// Run runs "adb -s <serial> <args>" and returns its stdout.
func (c Client) Run(ctx context.Context, args ...string) ([]byte, error) {
	return c.Runner.Run(ctx, Binary, c.args(args)...)
}

// Start starts "adb -s <serial> <args>" in the background.
func (c Client) Start(ctx context.Context, args ...string) (device.Process, error) {
	return c.Runner.Start(ctx, Binary, c.args(args)...)
}

// StartStreaming starts "adb -s <serial> <args>" in the background and writes
// its stdout and stderr to out, see device.CommandRunner.StartStreaming.
func (c Client) StartStreaming(ctx context.Context, out io.Writer, args ...string) (device.Process, error) {
	return c.Runner.StartStreaming(ctx, out, Binary, c.args(args)...)
}

func (c Client) args(args []string) []string {
	return append([]string{"-s", c.Serial}, args...)
}
