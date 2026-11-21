package androiddevice

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	defaultLogLines = 200
	defaultFollow   = 2000
)

// logcatTime is the time format of "logcat -v time" and of "logcat -T".
var logcatTime = regexp.MustCompile(`^\d\d-\d\d \d\d:\d\d:\d\d\.\d{3}$`)

// ReadLogs implements device.LogReader using a logcat dump.
func (d *Device) ReadLogs(lines int) ([]string, error) {
	if lines <= 0 {
		lines = defaultLogLines
	}
	out, err := d.run("logcat", "-d", "-v", "time", "-t", strconv.Itoa(lines))
	if err != nil {
		return nil, err
	}
	return splitLines(string(out)), nil
}

// ReadLogsSince implements device.LogFollower: the log lines since a time,
// at most the last lines of them.
func (d *Device) ReadLogsSince(since string, lines int) ([]string, error) {
	if !logcatTime.MatchString(since) {
		return nil, fmt.Errorf("invalid log time %q", since)
	}
	if lines <= 0 {
		lines = defaultFollow
	}
	// adb quotes each logcat argument for the device shell
	out, err := d.run("logcat", "-d", "-v", "time", "-T", since)
	if err != nil {
		return nil, err
	}
	result := splitLines(string(out))
	if len(result) > lines {
		result = result[len(result)-lines:]
	}
	return result, nil
}

// splitLines splits command output into lines; empty output gives nil.
func splitLines(out string) []string {
	result := strings.Split(strings.TrimRight(strings.ReplaceAll(out, "\r\n", "\n"), "\n"), "\n")
	if len(result) == 1 && result[0] == "" {
		return nil
	}
	return result
}
