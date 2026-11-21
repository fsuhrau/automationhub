// Package adb runs adb commands for one device through a device.CommandRunner.
package adb

import (
	"context"

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

func (c Client) args(args []string) []string {
	return append([]string{"-s", c.Serial}, args...)
}
