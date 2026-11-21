package hostinfo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Keys of the hardware parameters a device reports.
const (
	ParamModel  = "Device Model"
	ParamSerial = "Serial Number"
	ParamCPU    = "CPU"
	ParamGPU    = "GPU"
	ParamRAM    = "RAM"
	ParamDriver = "Driver"
)

const (
	hardwareData = "SPHardwareDataType"
	displaysData = "SPDisplaysDataType"
	gib          = 1024 * 1024 * 1024
)

// Parameters returns the requested hardware values; those that cannot be
// read are left out. Unknown keys are ignored.
func (h *Host) Parameters(ctx context.Context, keys ...string) map[string]string {
	probes := map[string]func(context.Context) (string, error){
		ParamModel:  h.ModelIdentifier,
		ParamSerial: h.SerialNumber,
		ParamCPU:    h.CPU,
		ParamGPU:    h.GPU,
		ParamRAM:    h.RAM,
		ParamDriver: h.GraphicsDrivers,
	}
	params := map[string]string{}
	for _, key := range keys {
		probe, ok := probes[key]
		if !ok {
			continue
		}
		if value, err := probe(ctx); err == nil {
			params[key] = value
		}
	}
	return params
}

// ModelIdentifier returns the Mac model, e.g. "MacBookPro18,1".
func (h *Host) ModelIdentifier(ctx context.Context) (string, error) {
	return h.firstProfileValue(ctx, hardwareData, "Model Identifier")
}

// SerialNumber returns the system serial number.
func (h *Host) SerialNumber(ctx context.Context) (string, error) {
	return h.firstProfileValue(ctx, hardwareData, "Serial Number (system)")
}

// GPU returns the chipset model.
func (h *Host) GPU(ctx context.Context) (string, error) {
	return h.firstProfileValue(ctx, displaysData, "Chipset Model")
}

// GraphicsDrivers returns the supported graphics APIs, comma separated.
func (h *Host) GraphicsDrivers(ctx context.Context) (string, error) {
	values, err := h.profileValues(ctx, displaysData, "Metal", "OpenGL")
	return strings.Join(values, ","), err
}

// CPU returns the CPU brand string.
func (h *Host) CPU(ctx context.Context) (string, error) {
	return h.macTrimmed(ctx, "sysctl", "-n", "machdep.cpu.brand_string")
}

// RAM returns the installed memory, e.g. "16 GB".
func (h *Host) RAM(ctx context.Context) (string, error) {
	value, err := h.macTrimmed(ctx, "sysctl", "-n", "hw.memsize")
	if err != nil {
		return "", err
	}
	bytes, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse memory size %q: %w", value, err)
	}
	return fmt.Sprintf("%d GB", bytes/gib), nil
}

func (h *Host) firstProfileValue(ctx context.Context, dataType, key string) (string, error) {
	values, err := h.profileValues(ctx, dataType, key)
	if err != nil {
		return "", err
	}
	if len(values) == 0 {
		return "", fmt.Errorf("%s: %w", key, ErrNotFound)
	}
	return values[0], nil
}

func (h *Host) profileValues(ctx context.Context, dataType string, keys ...string) ([]string, error) {
	out, err := h.macOutput(ctx, "system_profiler", dataType)
	if err != nil {
		return nil, err
	}
	return profileValues(out, keys), nil
}

// profileValues returns the value after the colon of every line that holds one of the keys.
func profileValues(profile string, keys []string) []string {
	var values []string
	for _, line := range strings.Split(profile, "\n") {
		if !containsAny(line, keys) {
			continue
		}
		if _, value, ok := strings.Cut(line, ":"); ok {
			values = append(values, strings.TrimSpace(strings.SplitN(value, ":", 2)[0]))
		}
	}
	return values
}

func containsAny(line string, keys []string) bool {
	for _, key := range keys {
		if strings.Contains(line, key) {
			return true
		}
	}
	return false
}
