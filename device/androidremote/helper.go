package androiddevice

import (
	"strings"

	"github.com/fsuhrau/automationhub/device"
)

func isDebuggablePackage(apkPath string) bool {
	cmd := device.NewCommand("aapt", "dump", "badging", apkPath)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "application-debuggable")
}
