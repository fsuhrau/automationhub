package androiddevice

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/fsuhrau/automationhub/tools/android"
)

// entryLabels maps the fields of an "adb devices -l" entry to the parameter
// names of the device, in display order.
func entryLabels() [][2]string {
	return [][2]string{
		{"device", "Device"},
		{"model", "Device Model"},
		{"transport_id", "Transport ID"},
		{"usb", "Device USB"},
	}
}

// UpdateDeviceInfos applies the fields of an "adb devices -l" entry. The
// first call also reads the device properties over adb.
func (d *Device) UpdateDeviceInfos(entry adbDeviceEntry) {
	first := d.Parameter == nil
	if first {
		d.Parameter = make(map[string]string)
	}
	d.Name = entry.Serial
	for _, label := range entryLabels() {
		if v := entry.Fields[label[0]]; v != "" {
			d.Parameter[label[1]] = v
		}
	}
	if first {
		d.readProperties()
	}
}

// readProperties reads API level, release, address and ABIs of the device.
func (d *Device) readProperties() {
	d.apiLevel = d.propertyInt("ro.build.version.sdk")
	d.OSVersion = d.property("ro.build.version.release")
	d.deviceIP, _ = d.lookupIP()
	d.Parameter["IP"] = d.deviceIP.String()
	d.Parameter["ABIs"] = d.property("ro.product.cpu.abilist")
	d.Parameter["API Level"] = strconv.FormatInt(d.apiLevel, 10)
}

// property reads a system property, empty when adb fails.
func (d *Device) property(name string) string {
	out, err := d.run("shell", "getprop", name)
	if err != nil {
		return ""
	}
	return strings.Trim(string(out), "\n\r")
}

// propertyInt reads a numeric system property, -1 when it is no number.
func (d *Device) propertyInt(name string) int64 {
	n, err := strconv.ParseInt(d.property(name), 10, 64)
	if err != nil {
		return -1
	}
	return n
}

// lookupIP reads the IPv4 address of the first network interface that has one.
func (d *Device) lookupIP() (net.IP, error) {
	var err error
	for _, nic := range []string{"wlan0", "eth0"} {
		var out []byte
		if out, err = d.run("shell", "ip", "-f", "inet", "addr", "show", nic); err == nil {
			return parseIP(out), nil
		}
	}
	return nil, fmt.Errorf("read ip address: %w", err)
}

// parseIP finds the first "inet" address in the output of "ip addr show".
func parseIP(out []byte) net.IP {
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		if m := android.IPLookupRegex.FindStringSubmatch(scanner.Text()); len(m) >= 3 {
			return net.ParseIP(m[1])
		}
	}
	return nil
}
