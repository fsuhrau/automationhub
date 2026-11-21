package web

import (
	"context"
	"fmt"
	"strings"

	"github.com/fsuhrau/automationhub/app"
	"github.com/sirupsen/logrus"
)

// versionCommand is the command that prints the version of the browser. Safari
// has no --version flag, its version is in the bundle's Info.plist.
func versionCommand(goos, browser, path string) (string, []string) {
	if goos == "darwin" && browser == "safari" {
		return "defaults", []string{"read", path + "/Contents/Info", "CFBundleShortVersionString"}
	}
	return path, []string{"--version"}
}

// parseBrowserVersion takes the last word of the version output, so
// "Google Chrome 126.0.1" gives "126.0.1".
func parseBrowserVersion(output string) string {
	output = strings.TrimSpace(output)
	if i := strings.LastIndex(output, " "); i != -1 {
		return output[i+1:]
	}
	return output
}

// readBrowserVersion stores the browser version as target and parameter.
func (d *Device) readBrowserVersion(ctx context.Context) {
	name, args := versionCommand(d.env.GOOS, d.browser, d.browserPath)
	out, err := d.env.Runner.Run(ctx, name, args...)
	if err != nil {
		logrus.Warnf("failed to get the version of %s: %v", d.browser, err)
		return
	}
	d.Target = parseBrowserVersion(string(out))
	d.Parameter["Browser Version"] = d.Target
}

// applicationURL is the start url with the session the app connects with.
func (d *Device) applicationURL(params *app.Parameter, sessionID, nodeURL string) string {
	return fmt.Sprintf("%s?sessionId=%s&nodeURL=%s&deviceId=%s", params.Web.StartURL, sessionID, nodeURL, d.DeviceID())
}
