package androiddevice

import "testing"

func TestParseADBDeviceLine(t *testing.T) {
	cases := []struct {
		line   string
		ok     bool
		serial string
		state  string
	}{
		{"List of devices attached", false, "", ""},
		{"* daemon started successfully", false, "", ""},
		{"", false, "", ""},
		{"RZCTA1YRJ5W            unauthorized usb:0-1.3 transport_id:2", true, "RZCTA1YRJ5W", "unauthorized"},
		{"emulator-5554          device product:sdk_gphone64_arm64 model:sdk_gphone64_arm64 device:emu64a transport_id:1", true, "emulator-5554", "device"},
		{"127.0.0.1:5555         offline transport_id:4", true, "127.0.0.1:5555", "offline"},
	}
	for _, c := range cases {
		entry, ok := parseADBDeviceLine(c.line)
		if ok != c.ok || entry.Serial != c.serial || entry.State != c.state {
			t.Errorf("parseADBDeviceLine(%q) = %+v, %v", c.line, entry, ok)
		}
	}
	entry, _ := parseADBDeviceLine("RZCTA1YRJ5W            unauthorized usb:0-1.3 transport_id:2")
	if entry.Fields["usb"] != "0-1.3" || entry.Fields["transport_id"] != "2" {
		t.Errorf("unexpected fields %v", entry.Fields)
	}
}

func TestDeviceLineStillMatchesRegex(t *testing.T) {
	line := "emulator-5554          device product:sdk_gphone64_arm64 model:sdk_gphone64_arm64 device:emu64a transport_id:1"
	if m := DeviceListRegex.FindAllStringSubmatch(line, -1); len(m) != 1 || m[0][1] != "emulator-5554" {
		t.Fatalf("regex no longer matches device lines: %v", m)
	}
}
