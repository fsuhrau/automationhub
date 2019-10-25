package androiddevice

import (
	"strings"

	"github.com/fsuhrau/automationhub/devices"
)

func isDebuggablePackage(apkPath string) bool {
	cmd := devices.NewCommand("aapt", "dump", "badging", apkPath)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "application-debuggable")
}
