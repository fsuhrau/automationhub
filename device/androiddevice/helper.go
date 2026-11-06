package androiddevice

import (
	"github.com/fsuhrau/automationhub/tools/androidsdk"
	"github.com/fsuhrau/automationhub/tools/exec"
	"strings"
)

func isDebuggablePackage(apkPath string) bool {
	aapt, err := androidsdk.AAPT()
	if err != nil {
		return false
	}
	cmd := exec.NewCommand(aapt, "dump", "badging", apkPath)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "application-debuggable")
}
