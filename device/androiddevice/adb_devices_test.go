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

func TestParseADBDeviceLineReported(t *testing.T) {
	// reported from a Galaxy S21 FE after authorization
	entry, ok := parseADBDeviceLine("RZCTA1YRJ5W            device usb:0-1.3 product:r9q2xeea model:SM_G990B2 device:r9q transport_id:3")
	if !ok || entry.Serial != "RZCTA1YRJ5W" || entry.State != adbStateDevice || entry.Fields["model"] != "SM_G990B2" || entry.Fields["transport_id"] != "3" {
		t.Fatalf("unexpected entry %+v", entry)
	}
}

func TestParseADBDeviceLineWithUSBPath(t *testing.T) {
	// newer adb versions print the usb port as a path (the old regex rejected it)
	entry, ok := parseADBDeviceLine("RZCTA1YRJ5W            device usb:0-1.3 product:a52qnsxx model:SM_A525F device:a52q transport_id:2")
	if !ok || entry.State != adbStateDevice || entry.Fields["model"] != "SM_A525F" || entry.Fields["usb"] != "0-1.3" || entry.Fields["product"] != "a52qnsxx" {
		t.Fatalf("unexpected entry %+v", entry)
	}
}
