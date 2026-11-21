package app

import (
	"fmt"
	"sort"
	"strings"

	"howett.net/plist"
)

// readInfoPlist fills the info from an app's Info.plist
func readInfoPlist(data []byte, info *BundleInfo) (map[string]interface{}, error) {
	content := plistDict{}
	if _, err := plist.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("Info.plist: %w", err) //nolint:staticcheck // ST1005: the message names the file and is shown as is
	}
	apple := appleInfoOf(content)
	info.Apple = apple
	info.Identifier = plistString(content, "CFBundleIdentifier")
	info.Version = plistString(content, "CFBundleShortVersionString")
	info.Build = plistString(content, "CFBundleVersion")
	info.DisplayName = plistString(content, "CFBundleDisplayName")
	if info.DisplayName == "" {
		info.DisplayName = apple.BundleName
	}
	return content, nil
}

func appleInfoOf(content plistDict) *AppleInfo {
	apple := &AppleInfo{
		Platform:             plistString(content, "DTPlatformName"),
		MinOS:                plistString(content, "MinimumOSVersion"),
		SDK:                  plistString(content, "DTSDKName"),
		Xcode:                xcodeVersion(plistString(content, "DTXcode")),
		XcodeBuild:           plistString(content, "DTXcodeBuild"),
		BuildMachineOS:       plistString(content, "BuildMachineOSBuild"),
		BundleName:           plistString(content, "CFBundleName"),
		Executable:           plistString(content, "CFBundleExecutable"),
		RequiredCapabilities: plistStrings(content, "UIRequiredDeviceCapabilities"),
		Orientations:         plistStrings(content, "UISupportedInterfaceOrientations"),
		BackgroundModes:      plistStrings(content, "UIBackgroundModes"),
		DeviceFamily:         deviceFamilyNames(plistStrings(content, "UIDeviceFamily")),
		URLSchemes:           urlSchemes(content),
		UsageDescriptions:    usageDescriptions(content),
	}
	if apple.MinOS == "" {
		apple.MinOS = plistString(content, "LSMinimumSystemVersion")
	}
	apple.FileSharing, _ = content["UIFileSharingEnabled"].(bool)
	return apple
}

func deviceFamilyNames(ids []string) []string {
	var names []string
	for _, id := range ids {
		names = append(names, deviceFamilyName(id))
	}
	return names
}

func deviceFamilyName(id string) string {
	switch id {
	case "1":
		return "iPhone"
	case "2":
		return "iPad"
	case "3":
		return "Apple TV"
	case "4":
		return "Apple Watch"
	case "6":
		return "Mac (Catalyst)"
	case "7":
		return "Apple Vision"
	}
	return id
}

func urlSchemes(content plistDict) []string {
	types, _ := content["CFBundleURLTypes"].([]interface{})
	var schemes []string
	for _, t := range types {
		if m, ok := t.(plistDict); ok {
			schemes = append(schemes, plistStrings(m, "CFBundleURLSchemes")...)
		}
	}
	return schemes
}

// usageDescriptions are the permission texts, sorted by key
func usageDescriptions(content plistDict) []UsageDescription {
	keys := make([]string, 0, len(content))
	for k := range content {
		if strings.HasSuffix(k, "UsageDescription") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var out []UsageDescription
	for _, k := range keys {
		out = append(out, UsageDescription{Key: k, Text: plistString(content, k)})
	}
	return out
}

// xcodeVersion turns DTXcode "1520" into 15.2
func xcodeVersion(v string) string {
	if len(v) < 3 {
		return v
	}
	major := strings.TrimLeft(v[:len(v)-2], "0")
	minor := v[len(v)-2 : len(v)-1]
	patch := v[len(v)-1:]
	if patch == "0" {
		return major + "." + minor
	}
	return major + "." + minor + "." + patch
}
