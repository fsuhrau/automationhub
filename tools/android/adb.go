package android

import (
	"bufio"
	"bytes"
	"github.com/fsuhrau/automationhub/tools/exec"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/fsuhrau/automationhub/app"
	"github.com/sirupsen/logrus"
)

const (
	ADB  = "adb"
	AAPT = "aapt"
)

var (
	IPLookupRegex = regexp.MustCompile(`\s+inet\s+([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})/[0-9]+\sbrd\s([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})\s+.*`)
)

func GetParameterString(deviceID, param string) string {
	cmd := exec.NewCommand(ADB, "-s", deviceID, "shell", "getprop", param)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.Trim(string(output), "\n")
}

func GetParameterInt(deviceID, param string) int64 {
	versionString := GetParameterString(deviceID, param)
	intParam, err := strconv.ParseInt(versionString, 10, 64)
	if err != nil {
		logrus.Errorf("Could not parse: %v", err)
		return -1
	}

	return intParam
}

func GetDeviceIP(deviceID string) (net.IP, error) {
	networkDevices := []string{
		"wlan0",
		"eth0",
	}
	var output []byte
	var err error
	for _, nd := range networkDevices {
		cmd := exec.NewCommand(ADB, "-s", deviceID, "shell", "ip", "-f", "inet", "addr", "show", nd)
		output, err = cmd.Output()
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		matches := IPLookupRegex.FindAllStringSubmatch(line, -1)
		if len(matches) < 1 {
			continue
		}
		if len(matches[0]) < 3 {
			continue
		}
		return net.ParseIP(matches[0][1]), nil
	}
	return nil, nil
}

func IsAppInstalled(deviceID string, params *app.Parameter) (bool, error) {
	cmd := exec.NewCommand(ADB, "-s", deviceID, "shell", "pm", "list", "packages")
	output, err := cmd.Output()
	return strings.Contains(string(output), params.Identifier), err
}
