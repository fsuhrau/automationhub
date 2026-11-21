package iossim

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/fsuhrau/automationhub/device"
)

var osVersionLookupRegex = regexp.MustCompile(`com.apple.CoreSimulator.SimRuntime.([a-z]+OS)-([0-9]+-[0-9]+)`)

// simctl runs `xcrun simctl` subcommands on a command runner.
type simctl struct {
	runner device.CommandRunner
}

func (s simctl) run(args ...string) ([]byte, error) {
	out, err := s.runner.Run(context.Background(), "xcrun", append([]string{"simctl"}, args...)...)
	if err != nil {
		return out, fmt.Errorf("simctl %s: %w", args[0], err)
	}
	return out, nil
}

// list returns the simulators `simctl list devices --json` reports.
func (s simctl) list() ([]simInfo, error) {
	out, err := s.run("list", "devices", "--json")
	if err != nil {
		return nil, err
	}
	return parseSimulators(out)
}

type simDevice struct {
	State string `json:"state"`
	Name  string `json:"name"`
	UDID  string `json:"udid"`
}

type simulatorDescriptions struct {
	Devices map[string][]simDevice `json:"devices"`
}

// simInfo is one simulator of the list output with its runtime resolved.
type simInfo struct {
	UDID, Name, OSName, OSVersion string
	State                         device.State
}

func parseSimulators(output []byte) ([]simInfo, error) {
	var resp simulatorDescriptions
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("parse simulator list: %w", err)
	}
	var sims []simInfo
	for runtime, devices := range resp.Devices {
		osName, osVersion, ok := parseRuntime(runtime)
		if !ok {
			continue
		}
		for _, d := range devices {
			sims = append(sims, simInfo{UDID: d.UDID, Name: d.Name, State: parseSimState(d.State), OSName: osName, OSVersion: osVersion})
		}
	}
	return sims, nil
}

// parseSimState maps the state simctl prints ("Booted", "Shutdown") to a
// device.State.
func parseSimState(state string) device.State {
	return device.ParseState("State" + state)
}

// parseRuntime splits a CoreSimulator runtime id into the OS name and version.
func parseRuntime(runtime string) (osName, osVersion string, ok bool) {
	subs := osVersionLookupRegex.FindStringSubmatch(runtime)
	if subs == nil {
		return "", "", false
	}
	osName = subs[1]
	if osName == "iOS" {
		osName = "iphonesimulator"
	}
	return osName, strings.ReplaceAll(subs[2], "-", "."), true
}
