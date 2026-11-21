package app

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseBadgingLines(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  AndroidInfo
	}{
		{"package", []string{"package: name='com.x' versionCode='3' versionName='1.0' compileSdkVersion='34' compileSdkVersionCodename='14' platformBuildVersionName='14'"},
			AndroidInfo{PackageName: "com.x", VersionCode: "3", VersionName: "1.0", CompileSdk: "34", CompileSdkCodename: "14", PlatformBuildVersion: "14"}},
		{"sdkVersion", []string{"sdkVersion:'24'"}, AndroidInfo{MinSdk: "24"}},
		{"minSdkVersion", []string{"minSdkVersion:'21'"}, AndroidInfo{MinSdk: "21"}},
		{"targetSdkVersion", []string{"targetSdkVersion:'34'"}, AndroidInfo{TargetSdk: "34"}},
		{"maxSdkVersion", []string{"maxSdkVersion:'35'"}, AndroidInfo{MaxSdk: "35"}},
		{"application-label", []string{"application-label:'Game'", "application-label-de:'Spiel'"}, AndroidInfo{Label: "Game"}},
		{"application label as fallback", []string{"application: label='App' icon=''"}, AndroidInfo{Label: "App"}},
		{"application label does not replace", []string{"application-label:'Game'", "application: label='App' icon=''"}, AndroidInfo{Label: "Game"}},
		{"launchable-activity first wins", []string{"launchable-activity: name='a.Main'", "launchable-activity: name='a.Other'"}, AndroidInfo{LaunchActivity: "a.Main"}},
		{"permissions", []string{
			"uses-permission: name='p.A'",
			"uses-permission-sdk-23: name='p.B' maxSdkVersion='28'",
			"uses-permission-sdk-m: name='p.C'",
			"uses-permission: name='p.A'",
			"uses-permission: name=''",
		}, AndroidInfo{Permissions: []AndroidPermission{{Name: "p.A"}, {Name: "p.B", MaxSdk: "28"}, {Name: "p.C"}}}},
		{"features", []string{
			"uses-feature-not-required: name='f.A'",
			"uses-feature: name='f.B'",
			"uses-feature: name='f.C' required='false'",
			"uses-feature: name='f.A'",
			"uses-feature: name=''",
		}, AndroidInfo{Features: []AndroidFeature{{Name: "f.A", Required: true}, {Name: "f.B", Required: true}, {Name: "f.C"}}}},
		{"uses-gl-es", []string{"uses-gl-es:'0x30001'"}, AndroidInfo{GLES: "3.1"}},
		{"native-code", []string{"native-code: 'arm64-v8a' 'x86'", "alt-native-code: 'armeabi-v7a'"}, AndroidInfo{ABIs: []string{"arm64-v8a", "x86", "armeabi-v7a"}}},
		{"locales", []string{"locales: '--_--' 'de' 'en'"}, AndroidInfo{Locales: []string{"de", "en"}}},
		{"densities", []string{"densities: '160' '640'"}, AndroidInfo{Densities: []string{"160", "640"}}},
		{"supports-screens", []string{"supports-screens: 'small' 'large'"}, AndroidInfo{Screens: []string{"small", "large"}}},
		{"flags", []string{"application-debuggable", "leanback-launcher", "testOnly='-1'"}, AndroidInfo{Debuggable: true, Leanback: true, TestOnly: true}},
		{"unknown lines are ignored", []string{"main", "other: 'x'", "application-icon-abc:'x'"}, AndroidInfo{}},
		{"indented and empty lines", []string{"", "  uses-feature: name='f.A'"}, AndroidInfo{Features: []AndroidFeature{{Name: "f.A", Required: true}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, _ := parseBadging([]byte(strings.Join(tt.lines, "\n")))
			if !reflect.DeepEqual(*info, tt.want) {
				t.Errorf("info = %+v, want %+v", *info, tt.want)
			}
		})
	}
}

func TestParseBadgingIcons(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  map[int]string
	}{
		{"densities", []string{"application-icon-160:'res/a.png'", "application-icon-640:'res/b.png'"}, map[int]string{160: "res/a.png", 640: "res/b.png"}},
		{"application icon is density 0, first wins", []string{"application: label='' icon='res/x.xml'", "application: label='' icon='res/y.xml'"}, map[int]string{0: "res/x.xml"}},
		{"application without icon", []string{"application: label='A' icon=''"}, map[int]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, icons := parseBadging([]byte(strings.Join(tt.lines, "\n")))
			if !reflect.DeepEqual(icons, tt.want) {
				t.Errorf("icons = %v, want %v", icons, tt.want)
			}
		})
	}
}

func TestBadgingLine(t *testing.T) {
	l := newBadgingLine("k", " 'a' 'b\\'c' x='1'")
	if got := l.value(); got != "a" {
		t.Errorf("value = %q", got)
	}
	if got := l.list(); !reflect.DeepEqual(got, []string{"a", "b\\'c", "1"}) {
		t.Errorf("list = %q", got)
	}
	if got := l.attrs["x"]; got != "1" {
		t.Errorf("attrs = %v", l.attrs)
	}
	bare := newBadgingLine("k", " 0x30000 ")
	if got := bare.value(); got != "0x30000" {
		t.Errorf("value without quotes = %q", got)
	}
}

func TestGLESVersionKeepsUnparsable(t *testing.T) {
	if got := glesVersion("bogus"); got != "bogus" {
		t.Errorf("glesVersion = %q", got)
	}
}

func TestFeatureSetIsEmptyByDefault(t *testing.T) {
	var f featureSet
	if got := f.list(); got != nil {
		t.Errorf("list = %v", got)
	}
}
