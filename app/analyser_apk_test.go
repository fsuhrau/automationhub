package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const apkBadging = `package: name='com.example.game' versionCode='7' versionName='1.2'
application-label:'Game'
application-icon-640:'res/mipmap-anydpi/ic.xml'
launchable-activity: name='com.example.Main'
`

func apkAnalyser(t *testing.T, files map[string][]byte, badging func(string) ([]byte, error)) *analyser {
	t.Helper()
	path := filepath.Join(t.TempDir(), "game.apk")
	writeZip(t, path, files)
	a := NewAnalyser(path)
	a.badging = badging
	return a
}

func fixedBadging(output string) func(string) ([]byte, error) {
	return func(string) ([]byte, error) { return []byte(output), nil }
}

func TestAnalyseAPK(t *testing.T) {
	a := apkAnalyser(t, map[string][]byte{
		"lib/arm64-v8a/libunity.so": []byte("unity"), "res/mipmap-xxhdpi/ic.png": []byte("png"), "res/mipmap-anydpi/ic.xml": []byte("xml"),
		"META-INF/CERT.RSA": v1Signature(t, "Legacy"),
	}, fixedBadging(apkBadging))
	if err := a.AnalyseFile(); err != nil {
		t.Fatal(err)
	}
	p := a.GetParameter()
	if p.Platform != "android" || p.Identifier != "com.example.game" || p.Version != "1.2" || p.Name != "Game" {
		t.Errorf("parameter = %+v", p)
	}
	if p.App.Android == nil || p.App.Android.LaunchActivity != "com.example.Main" || p.App.Size == 0 || len(p.App.Hash) != 40 {
		t.Errorf("app = %+v", p.App)
	}
	info := p.Info
	if info.Format != "apk" || info.Build != "7" || info.DisplayName != "Game" || info.Icon != "res/mipmap-xxhdpi/ic.png" {
		t.Errorf("info = %+v", info)
	}
	if info.Engine == nil || info.Engine.Name != "Unity" || info.Contents == nil || info.Contents.Files != 4 {
		t.Errorf("engine = %+v contents = %+v", info.Engine, info.Contents)
	}
	if info.Signing == nil || len(info.Signing.Schemes) != 1 || len(info.Warnings) != 0 {
		t.Errorf("signing = %+v warnings = %v", info.Signing, info.Warnings)
	}
}

func TestAnalyseAPKWithoutLaunchActivityOrSignature(t *testing.T) {
	a := apkAnalyser(t, map[string][]byte{"classes.dex": []byte("dex")}, fixedBadging("package: name='com.x'\n"))
	if err := a.ReadInfo(); err != nil {
		t.Fatal(err)
	}
	p := a.GetParameter()
	if p.App.Android != nil || p.Info.Signing != nil || p.Info.Engine != nil || p.App.Hash != "" {
		t.Errorf("parameter = %+v info = %+v", p, p.Info)
	}
}

func TestAnalyseAPKWarnsWhenItIsNoZip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "game.apk")
	writeFile(t, path, []byte("not a zip"))
	a := NewAnalyser(path)
	a.badging = fixedBadging("package: name='com.x'\n")
	if err := a.ReadInfo(); err != nil {
		t.Fatal(err)
	}
	if w := a.GetParameter().Info.Warnings; len(w) != 1 || !strings.HasPrefix(w[0], "unable to read the apk as zip: ") {
		t.Errorf("warnings = %v", w)
	}
}

func TestAnalyseAPKErrors(t *testing.T) {
	boom := errors.New("aapt failed")
	a := apkAnalyser(t, map[string][]byte{"a": nil}, func(string) ([]byte, error) { return nil, boom })
	if err := a.AnalyseFile(); !errors.Is(err, boom) {
		t.Errorf("err = %v", err)
	}
	missing := NewAnalyser(filepath.Join(t.TempDir(), "none.apk"))
	missing.badging = fixedBadging("package: name='com.x'\n")
	if err := missing.ReadInfo(); err == nil {
		t.Error("a missing apk is an error")
	}
}

func TestAnalyseAPKCannotOpenAnUnreadableFile(t *testing.T) {
	a := apkAnalyser(t, map[string][]byte{"a": nil}, fixedBadging("package: name='com.x'\n"))
	if err := os.Chmod(a.appPath, 0); err != nil {
		t.Fatal(err)
	}
	if f, err := os.Open(a.appPath); err == nil {
		_ = f.Close()
		t.Skip("the file is readable anyway, running as root")
	}
	if err := a.ReadInfo(); err == nil {
		t.Error("an unreadable apk is an error")
	}
}

func TestApplySigning(t *testing.T) {
	tests := []struct {
		name        string
		signing     *Signing
		err         error
		wantSigning bool
		wantWarning bool
	}{
		{"signed", &Signing{Schemes: []string{"v2"}}, nil, true, false},
		{"certificates only", &Signing{Certificates: []Certificate{{}}}, nil, true, false},
		{"unsigned", &Signing{}, nil, false, false},
		{"nothing read", nil, nil, false, false},
		{"partly read", &Signing{Schemes: []string{"v1"}}, errors.New("bad block"), true, true},
		{"failed", &Signing{}, errors.New("bad block"), false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &BundleInfo{}
			applySigning(info, tt.signing, tt.err)
			if (info.Signing != nil) != tt.wantSigning || (len(info.Warnings) == 1) != tt.wantWarning {
				t.Errorf("signing = %+v warnings = %v", info.Signing, info.Warnings)
			}
		})
	}
}

func TestRunAAPTBadging(t *testing.T) {
	dir := t.TempDir()
	script := "#!/bin/sh\necho \"package: name='from.aapt' $3\"\n"
	if err := os.WriteFile(filepath.Join(dir, "aapt"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	out, err := runAAPTBadging("game.apk")
	if err != nil || !strings.Contains(string(out), "name='from.aapt' game.apk") {
		t.Errorf("out = %q, err = %v", out, err)
	}
}

func TestRunAAPTBadgingWithoutAAPT(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ANDROID_HOME", "")
	t.Setenv("ANDROID_SDK_ROOT", "")
	if _, err := runAAPTBadging("game.apk"); err == nil {
		t.Error("no error without aapt")
	}
}
