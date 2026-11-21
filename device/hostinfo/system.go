package hostinfo

import (
	"context"
	"fmt"
	"strings"
)

// ErrNotFound is returned when a command ran but its output lacks the value.
var ErrNotFound = fmt.Errorf("value not found in command output") //nolint:gochecknoglobals // sentinel error

// OSName returns the product name on macOS and the GOOS name elsewhere.
func (h *Host) OSName(ctx context.Context) (string, error) {
	if h.goos != darwin {
		return h.goos, nil
	}
	return h.macTrimmed(ctx, "sw_vers", "-productName")
}

// OSVersion returns the product version; only macOS is supported.
func (h *Host) OSVersion(ctx context.Context) (string, error) {
	return h.macTrimmed(ctx, "sw_vers", "-productVersion")
}

// HostName returns the machine name.
func (h *Host) HostName(ctx context.Context) (string, error) {
	if h.goos == darwin {
		return h.macTrimmed(ctx, "sysctl", "-n", "kern.hostname")
	}
	out, err := h.output(ctx, "hostname")
	return strings.TrimSpace(out), err
}

// HardwareUUID returns the IOPlatformUUID; only macOS is supported.
func (h *Host) HardwareUUID(ctx context.Context) (string, error) {
	out, err := h.macOutput(ctx, "ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	if err != nil {
		return "", err
	}
	return parseHardwareUUID(out)
}

// parseHardwareUUID finds the line `"IOPlatformUUID" = "<uuid>"`.
func parseHardwareUUID(ioreg string) (string, error) {
	for _, line := range strings.Split(ioreg, "\n") {
		if !strings.Contains(line, "IOPlatformUUID") {
			continue
		}
		if parts := strings.Split(line, "\""); len(parts) > 3 {
			return parts[3], nil
		}
	}
	return "", fmt.Errorf("hardware uuid: %w", ErrNotFound)
}
