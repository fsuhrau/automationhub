package androiddevice

import (
	"reflect"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
)

const ipOutput = "3: wlan0: <UP>\n    inet 192.168.1.5/24 brd 192.168.1.255 scope global wlan0\n"

func TestUpdateDeviceInfosReadsPropertiesOnce(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 shell getprop ro.build.version.sdk", []byte("30\r\n"), nil)
	runner.Script("adb -s S1 shell getprop ro.build.version.release", []byte("11\n"), nil)
	runner.Script("adb -s S1 shell ip -f inet addr show wlan0", []byte(ipOutput), nil)
	runner.Script("adb -s S1 shell getprop ro.product.cpu.abilist", []byte("arm64-v8a,armeabi\n"), nil)
	d := newTestDevice(runner)
	entry := adbDeviceEntry{Serial: "S1", State: adbStateDevice, Fields: map[string]string{
		"device": "r9q", "model": "SM_G990B2", "transport_id": "3", "usb": "0-1.3", "product": "x"}}

	d.UpdateDeviceInfos(entry)
	d.UpdateDeviceInfos(entry) // the second call must not run adb again

	want := map[string]string{
		"Device": "r9q", "Device Model": "SM_G990B2", "Transport ID": "3", "Device USB": "0-1.3",
		"IP": "192.168.1.5", "ABIs": "arm64-v8a,armeabi", "API Level": "30",
	}
	if !reflect.DeepEqual(d.Parameter, want) {
		t.Fatalf("parameter = %v, want %v", d.Parameter, want)
	}
	if d.OSVersion != "11" || d.apiLevel != 30 || d.Name != "S1" || d.DeviceIP().String() != "192.168.1.5" {
		t.Fatalf("version %q api %d name %q ip %v", d.OSVersion, d.apiLevel, d.Name, d.DeviceIP())
	}
	assertCommands(t, runner,
		"adb -s S1 shell getprop ro.build.version.sdk",
		"adb -s S1 shell getprop ro.build.version.release",
		"adb -s S1 shell ip -f inet addr show wlan0",
		"adb -s S1 shell getprop ro.product.cpu.abilist")
}

func TestUpdateDeviceInfosWhenAdbFails(t *testing.T) {
	d := newTestDevice(&testkit.FakeRunner{}) // every command fails
	d.UpdateDeviceInfos(adbDeviceEntry{Serial: "S1", Fields: map[string]string{}})
	if d.apiLevel != -1 || d.OSVersion != "" || d.Parameter["IP"] != "<nil>" || d.Parameter["API Level"] != "-1" {
		t.Fatalf("api %d version %q parameter %v", d.apiLevel, d.OSVersion, d.Parameter)
	}
}

func TestLookupIPFallsBackToEthernet(t *testing.T) {
	runner := &testkit.FakeRunner{}
	runner.Script("adb -s S1 shell ip -f inet addr show eth0", []byte("    inet 10.0.0.7/8 brd 10.255.255.255 scope global eth0\n"), nil)
	ip, err := newTestDevice(runner).lookupIP()
	if err != nil || ip.String() != "10.0.0.7" {
		t.Fatalf("ip %v err %v", ip, err)
	}
	assertCommands(t, runner, "adb -s S1 shell ip -f inet addr show wlan0", "adb -s S1 shell ip -f inet addr show eth0")
}

func TestLookupIPFailsWithoutInterface(t *testing.T) {
	if _, err := newTestDevice(&testkit.FakeRunner{}).lookupIP(); err == nil {
		t.Fatal("expected an error")
	}
}

func TestParseIP(t *testing.T) {
	if ip := parseIP([]byte(ipOutput)); ip.String() != "192.168.1.5" {
		t.Fatalf("ip = %v", ip)
	}
	if ip := parseIP([]byte("no address here\n")); ip != nil {
		t.Fatalf("ip = %v", ip)
	}
}
