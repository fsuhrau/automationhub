package macos

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// getSwVersValue runs a sw_vers command and returns the output as a string
func getSwVersValue(key string) (string, error) {
	cmd := exec.Command("sw_vers", "-"+key)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// getSysctlValue runs a sysctl command and returns the output as a string
func getSysctlValue(name string) (string, error) {
	cmd := exec.Command("sysctl", "-n", name)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// getSystemProfilerValue runs a system_profiler command and returns the output as a string
func getSystemProfilerValue(dataType, key string) (string, error) {
	cmd := exec.Command("system_profiler", dataType)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	output := out.String()
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, key) {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}
	return "", nil
}

// GetDeviceName returns the device name
func GetDeviceName() (string, error) {
	return getSysctlValue("kern.hostname")
}

// GetOSName returns the OS name
func GetOSName() (string, error) {
	return getSwVersValue("productName")
}

// GetOSVersion returns the OS version
func GetOSVersion() (string, error) {
	return getSwVersValue("productVersion")
}

func GetHardwareUUID() (string, error) {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	output := out.String()
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.Split(line, "\"")
			if len(parts) > 3 {
				return parts[3], nil
			}
		}
	}
	return "", fmt.Errorf("hardware UUID not found")
}

// GetSerialNumber returns the serial number of the Mac
func GetSerialNumber() (string, error) {
	return getSystemProfilerValue("SPHardwareDataType", "Serial Number (system)")
}

// GetModelNumber returns the model number of the Mac
func GetModelNumber() (string, error) {
	return getSystemProfilerValue("SPHardwareDataType", "Model Identifier")
}
