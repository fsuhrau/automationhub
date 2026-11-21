package app

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadInfoPlist(t *testing.T) {
	data := mustPlist(t, map[string]interface{}{
		"CFBundleIdentifier": "com.x", "CFBundleShortVersionString": "1.2", "CFBundleVersion": "9", "CFBundleName": "Name",
		"CFBundleExecutable": "Exe", "DTPlatformName": "iphoneos", "DTSDKName": "iphoneos17", "DTXcode": "1520", "DTXcodeBuild": "15C",
		"BuildMachineOSBuild": "23A", "LSMinimumSystemVersion": "12.0", "UIFileSharingEnabled": true,
		"UIDeviceFamily":                   []int{1, 2, 5},
		"UIRequiredDeviceCapabilities":     []string{"arm64"},
		"UISupportedInterfaceOrientations": []string{"UIInterfaceOrientationPortrait"},
		"UIBackgroundModes":                []string{"audio"},
		"CFBundleURLTypes": []interface{}{
			map[string]interface{}{"CFBundleURLSchemes": []string{"game", "game2"}}, "junk",
		},
		"NSMicrophoneUsageDescription": "Talk", "NSCameraUsageDescription": "Scan", "Other": "x",
	})
	info := &BundleInfo{}
	content, err := readInfoPlist(data, info)
	if err != nil {
		t.Fatal(err)
	}
	if content["Other"] != "x" {
		t.Errorf("content = %v", content)
	}
	if info.Identifier != "com.x" || info.Version != "1.2" || info.Build != "9" || info.DisplayName != "Name" {
		t.Errorf("info = %+v", info)
	}
	want := &AppleInfo{
		Platform: "iphoneos", MinOS: "12.0", SDK: "iphoneos17", Xcode: "15.2", XcodeBuild: "15C", BuildMachineOS: "23A",
		BundleName: "Name", Executable: "Exe", FileSharing: true,
		DeviceFamily:         []string{"iPhone", "iPad", "5"},
		RequiredCapabilities: []string{"arm64"}, Orientations: []string{"UIInterfaceOrientationPortrait"}, BackgroundModes: []string{"audio"},
		URLSchemes:        []string{"game", "game2"},
		UsageDescriptions: []UsageDescription{{"NSCameraUsageDescription", "Scan"}, {"NSMicrophoneUsageDescription", "Talk"}},
	}
	if !reflect.DeepEqual(info.Apple, want) {
		t.Errorf("apple = %+v, want %+v", info.Apple, want)
	}
}

func TestReadInfoPlistPrefersMinimumOSVersionAndDisplayName(t *testing.T) {
	data := mustPlist(t, map[string]interface{}{
		"MinimumOSVersion": "15.0", "LSMinimumSystemVersion": "12.0", "CFBundleDisplayName": "Shown", "CFBundleName": "Name",
	})
	info := &BundleInfo{}
	if _, err := readInfoPlist(data, info); err != nil {
		t.Fatal(err)
	}
	if info.Apple.MinOS != "15.0" || info.DisplayName != "Shown" || info.Apple.FileSharing {
		t.Errorf("info = %+v apple = %+v", info, info.Apple)
	}
}

func TestReadInfoPlistRejectsGarbage(t *testing.T) {
	if _, err := readInfoPlist([]byte("not a plist"), &BundleInfo{}); err == nil || !strings.HasPrefix(err.Error(), "Info.plist: ") {
		t.Errorf("err = %v", err)
	}
}

func TestDeviceFamilyName(t *testing.T) {
	for id, want := range map[string]string{
		"1": "iPhone", "2": "iPad", "3": "Apple TV", "4": "Apple Watch", "6": "Mac (Catalyst)", "7": "Apple Vision", "9": "9",
	} {
		if got := deviceFamilyName(id); got != want {
			t.Errorf("deviceFamilyName(%s) = %s, want %s", id, got, want)
		}
	}
}

func TestURLSchemesWithoutTypes(t *testing.T) {
	if got := urlSchemes(plistDict{}); got != nil {
		t.Errorf("schemes = %v", got)
	}
}
