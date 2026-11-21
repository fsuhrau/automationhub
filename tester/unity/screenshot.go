package unity

import (
	"context"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	tester_action "github.com/fsuhrau/automationhub/tester/action"
)

// screenshotTimeout bounds how long the app gets to take a screenshot.
const screenshotTimeout = 10 * time.Second

// captureScreenShot saves the screen of the device for a failed test.
func (s *strategy) captureScreenShot(ctx context.Context, tc testCase) {
	data := s.screenshot(ctx, tc)
	if data == nil {
		return
	}
	nameData := []byte(fmt.Sprintf("%d%s%s%s", time.Now().UnixNano(), s.runner.TestRun.SessionID, tc.dev.Device.DeviceID(), tc.task.Method))
	fileName := fmt.Sprintf("%x.png", sha1.Sum(nameData))
	if err := os.WriteFile(filepath.Join(s.runner.DataDir, fileName), data, os.ModePerm); err != nil {
		s.runner.LogError("Save screenshot failed: %v", err)
		return
	}
	tc.dev.Device.Data("screen", fileName)
}

// deviceScreenshot is the native screenshot of the device; it is nil when the
// device has no Screenshotter role or fails.
func deviceScreenshot(dev device.Device) []byte {
	shooter, ok := device.As[device.Screenshotter](dev)
	if !ok {
		return nil
	}
	data, _, _, _ := shooter.GetScreenshot()
	return data
}

// screenshot asks the device and, when it can't, the app for the screen.
func (s *strategy) screenshot(ctx context.Context, tc testCase) []byte {
	if data := deviceScreenshot(tc.dev.Device); data != nil {
		return data
	}
	var get action.GetScreenshot
	if err := tester_action.NewExecutor(s.runner.Devices).Execute(ctx, tester_action.Call{Device: tc.dev.Device, Action: &get, Timeout: screenshotTimeout}); err != nil {
		s.runner.LogError("Take screenshot failed: %v", err)
		return nil
	}
	return get.ScreenshotData()
}
