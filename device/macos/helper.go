package macos

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	SW_VERS         = "sw_vers"
	SYSCTL          = "sysctl"
	SYSTEM_PROFILER = "system_profiler"
)

var cache map[string]string

func init() {
	cache = make(map[string]string)
}

func cacheKey(cmd, param string) string {
	return fmt.Sprintf("%s_%s", cmd, param)
}

// getSwVersValue runs a sw_vers command and returns the output as a string
func getSwVersValue(key string) (string, error) {
	ck := cacheKey(SW_VERS, key)
	if cacheValue, ok := cache[ck]; ok {
		return cacheValue, nil
	}

	cmd := exec.Command(SW_VERS, "-"+key)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	result := strings.TrimSpace(out.String())
	cache[ck] = result
	return result, nil
}

// getSysctlValue runs a sysctl command and returns the output as a string
func getSysctlValue(name string) (string, error) {
	ck := cacheKey(SYSCTL, name)
	if cacheValue, ok := cache[ck]; ok {
		return cacheValue, nil
	}

	cmd := exec.Command(SYSCTL, "-n", name)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(out.String())
	cache[ck] = result
	return result, nil
}

// getSystemProfilerValue runs a system_profiler command and returns the output as a string
func getSystemProfilerValue(dataType string, keys []string) ([]string, error) {
	ck := cacheKey(SYSTEM_PROFILER, dataType)
	cacheValue, ok := cache[ck]
	if !ok {
		cmd := exec.Command("system_profiler", dataType)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return nil, err
		}
		cacheValue = out.String()
		cache[ck] = cacheValue
	}

	lines := strings.Split(cacheValue, "\n")
	var result []string
	for _, line := range lines {
		containsKey := false
		for _, key := range keys {
			if strings.Contains(line, key) {
				containsKey = true
				break
			}
		}
		if containsKey {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				result = append(result, strings.TrimSpace(parts[1]))
			}
		}
	}
	return result, nil
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
	values, err := getSystemProfilerValue("SPHardwareDataType", []string{"Serial Number (system)"})
	if err != nil {
		return "", err
	}

	if len(values) > 0 {
		return values[0], nil
	}
	return "", fmt.Errorf("serial number not found")
}

// GetModelNumber returns the model number of the Mac
func GetModelNumber() (string, error) {
	values, err := getSystemProfilerValue("SPHardwareDataType", []string{"Model Identifier"})
	if err != nil {
		return "", err
	}

	if len(values) > 0 {
		return values[0], nil
	}
	return "", fmt.Errorf("model not found")
}

func GetCPUInfo() (string, error) {
	return getSysctlValue("machdep.cpu.brand_string")
}

func GetGPUInfo() (string, error) {
	values, err := getSystemProfilerValue("SPDisplaysDataType", []string{"Chipset Model"})
	if err != nil {
		return "", err
	}

	if len(values) > 0 {
		return values[0], nil
	}

	return "", fmt.Errorf("chipset not found")
}

func GetRAMInfo() (string, error) {
	value, err := getSysctlValue("hw.memsize")
	if err != nil {
		return "", err
	}

	ramBytes, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return "", err
	}
	ramGB := ramBytes / (1024 * 1024 * 1024)
	return fmt.Sprintf("%d GB", ramGB), nil
}

func GetSupportedGraphicsEngines() ([]string, error) {
	return getSystemProfilerValue("SPDisplaysDataType", []string{"Metal", "OpenGL"})
}
