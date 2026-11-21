package unityeditor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

const (
	StartupTimeout = 2 * time.Minute
	pollInterval   = 5 * time.Second
	// instanceFileName is written into the project once the editor is up.
	instanceFileName = ".hub"
)

var errStartTimeout = errors.New("device didn't start in time")

func (m *Handler) StartDevice(deviceID string) error {
	dev, ok := m.Get(deviceID)
	if !ok {
		return device.DeviceNotFoundError
	}
	projectDir := projectDirOf(dev)
	instanceFile := filepath.Join(projectDir, instanceFileName)
	_ = os.Remove(instanceFile)

	deadline := m.clock.Now().Add(StartupTimeout)
	if err := m.launch(dev, projectDir); err != nil {
		return err
	}
	if err := m.waitFor(deadline, func() bool { return fileExists(instanceFile) }); err != nil {
		return err
	}
	return m.waitFor(deadline, func() bool { return dev.DeviceState() == device.StateBooted })
}

func projectDirOf(dev *Device) string {
	if config := dev.GetConfig(); config != nil {
		return config.GetAttribute("projectDir")
	}
	return ""
}

// launch starts the editor process for the project and follows its log.
func (m *Handler) launch(dev *Device, projectDir string) error {
	version, err := GetUsedUnityVersion(projectDir)
	if err != nil {
		return err
	}
	logFile := filepath.Join(projectDir, fmt.Sprintf("automation_hub_unity_%d.log", m.clock.Now().UnixNano()))
	editor := editorBinary(m.goos, m.managerCfg.UnityPath, version)
	process, err := m.runner.Start(context.Background(), editor, m.editorArgs(projectDir, logFile)...)
	if err != nil {
		return fmt.Errorf("start unity editor: %w", err)
	}
	dev.setLaunch(process, logFile)
	dev.UnityLogStartListening()
	return nil
}

func (m *Handler) editorArgs(projectDir, logFile string) []string {
	return []string{
		"-buildTarget", m.managerCfg.UnityBuildTarget,
		"-projectPath", projectDir,
		"-overrideProfile", "automation",
		"-executeMethod", "AutomationLoader.LoadSceneAndConnect",
		"-logFile", logFile,
		"-debugCodeOptimization",
		"--ump-channel-service-on-startup",
	}
}

// waitFor polls ready until it is true or the deadline passed.
func (m *Handler) waitFor(deadline time.Time, ready func() bool) error {
	for !ready() {
		if m.clock.Now().After(deadline) {
			return errStartTimeout
		}
		<-m.clock.After(pollInterval)
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (d *Device) setLaunch(process device.Process, logFile string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.process = process
	d.instanceLogFile = logFile
}

// stopProcess kills the editor process, if there is one.
func (d *Device) stopProcess() error {
	d.mu.Lock()
	process := d.process
	d.mu.Unlock()
	if process == nil {
		return nil
	}
	if err := process.Kill(); err != nil {
		return fmt.Errorf("kill unity editor: %w", err)
	}
	_ = process.Wait()
	d.setLaunch(nil, d.logFile())
	return nil
}

func (d *Device) logFile() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.instanceLogFile
}
