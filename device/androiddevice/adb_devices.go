package androiddevice

import "strings"

const (
	adbStateDevice       = "device"
	adbStateUnauthorized = "unauthorized"
)

// adbDeviceEntry is one line of "adb devices -l".
type adbDeviceEntry struct {
	Serial string
	// State is device, unauthorized, offline, recovery, ... ("no permissions"
	// is reported as "no")
	State  string
	Fields map[string]string
}

// parseADBDeviceLine parses lines like
//
//	emulator-5554    device product:sdk model:Pixel device:emu64 transport_id:1
//	RZCTA1YRJ5W      unauthorized usb:0-1.3 transport_id:2
func parseADBDeviceLine(line string) (adbDeviceEntry, bool) {
	fields := strings.Fields(line)
	if len(fields) < 2 || strings.HasPrefix(line, "List of devices") || strings.HasPrefix(line, "*") {
		return adbDeviceEntry{}, false
	}
	entry := adbDeviceEntry{Serial: fields[0], State: fields[1], Fields: map[string]string{}}
	for _, f := range fields[2:] {
		if k, v, ok := strings.Cut(f, ":"); ok {
			entry.Fields[k] = v
		}
	}
	return entry, true
}
