package app

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"compress/zlib"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"howett.net/plist"
)

const sampleBadging = `package: name='com.example.game' versionCode='393' versionName='1.58.0' platformBuildVersionName='14' platformBuildVersionCode='34' compileSdkVersion='34' compileSdkVersionCodename='14'
sdkVersion:'24'
targetSdkVersion:'34'
uses-permission: name='android.permission.INTERNET'
uses-permission: name='android.permission.WRITE_EXTERNAL_STORAGE' maxSdkVersion='28'
uses-permission: name='android.permission.INTERNET'
application-label:'Example Game'
application-label-de:'Beispielspiel'
application-icon-160:'res/mipmap-anydpi-v26/app_icon.xml'
application-icon-640:'res/mipmap-anydpi-v26/app_icon.xml'
application: label='Example Game' icon='res/mipmap-anydpi-v26/app_icon.xml'
application-debuggable
launchable-activity: name='com.unity3d.player.UnityPlayerActivity'  label='Example Game' icon=''
uses-gl-es:'0x30000'
uses-feature-not-required: name='android.hardware.vulkan.version'
uses-feature: name='android.hardware.touchscreen'
feature-group: label=''
  uses-feature: name='android.hardware.faketouch'
  uses-feature: name='android.hardware.touchscreen' required='false'
main
supports-screens: 'small' 'normal' 'large' 'xlarge'
supports-any-density: 'true'
locales: '--_--' 'de' 'en'
densities: '160' '640' '65535'
native-code: 'arm64-v8a'
alt-native-code: 'armeabi-v7a'
`

func TestParseBadging(t *testing.T) {
	info, icons := parseBadging([]byte(sampleBadging))
	check := func(name, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	check("package", info.PackageName, "com.example.game")
	check("versionCode", info.VersionCode, "393")
	check("versionName", info.VersionName, "1.58.0")
	check("compileSdk", info.CompileSdk, "34")
	check("minSdk", info.MinSdk, "24")
	check("targetSdk", info.TargetSdk, "34")
	check("label", info.Label, "Example Game")
	check("launch", info.LaunchActivity, "com.unity3d.player.UnityPlayerActivity")
	check("gles", info.GLES, "3.0")
	if len(info.Permissions) != 2 || info.Permissions[1].MaxSdk != "28" {
		t.Errorf("permissions = %+v", info.Permissions)
	}
	if len(info.Features) != 3 || !info.Features[1].Required || info.Features[0].Required {
		t.Errorf("features = %+v", info.Features)
	}
	if len(info.ABIs) != 2 || len(info.Locales) != 2 || len(info.Densities) != 3 || len(info.Screens) != 4 {
		t.Errorf("lists: abis %v locales %v densities %v screens %v", info.ABIs, info.Locales, info.Densities, info.Screens)
	}
	if !info.Debuggable {
		t.Error("not debuggable")
	}
	if icons[640] != "res/mipmap-anydpi-v26/app_icon.xml" {
		t.Errorf("icons = %v", icons)
	}
}

func TestVersions(t *testing.T) {
	for in, want := range map[string]string{"1520": "15.2", "1301": "13.0.1", "0941": "9.4.1"} {
		if got := xcodeVersion(in); got != want {
			t.Errorf("xcodeVersion(%s) = %s, want %s", in, got, want)
		}
	}
	if got := glesVersion("0x00030002"); got != "3.2" {
		t.Errorf("glesVersion = %s", got)
	}
	head := "\x00\x00\x00\x00\x00\x00\x00\x16\x00\x00\x00\x002022.3.20f1\x00"
	if got := unityVersionRegex.FindString(head); got != "2022.3.20f1" {
		t.Errorf("unity version = %q", got)
	}
}

func TestGroupOf(t *testing.T) {
	for in, want := range map[string]string{
		"classes.dex":                                        "classes.dex",
		"lib/arm64-v8a/libunity.so":                          "lib/arm64-v8a/",
		"assets/bin/Data/Managed/x.dll":                      "assets/bin/Data/",
		"res/drawable/x.png":                                 "res/drawable/",
		"Frameworks/UnityFramework.framework/UnityFramework": "Frameworks/UnityFramework.framework/",
		"Data/Raw/x.bundle":                                  "Data/Raw/",
		"_CodeSignature/CodeResources":                       "_CodeSignature/",
	} {
		if got := groupOf(in); got != want {
			t.Errorf("groupOf(%s) = %s, want %s", in, got, want)
		}
	}
}

func writeZip(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := zip.NewWriter(f)
	for name, data := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		fw.Write(data)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func mustPlist(t *testing.T, v interface{}) []byte {
	t.Helper()
	data, err := plist.Marshal(v, plist.XMLFormat)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestAnalyseIPA(t *testing.T) {
	dir := t.TempDir()
	ipa := filepath.Join(dir, "Game.ipa")
	profile := mustPlist(t, map[string]interface{}{
		"Name": "Game Dev", "TeamName": "Team", "TeamIdentifier": []string{"ABCDE12345"},
		"CreationDate": time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "ExpirationDate": time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		"ProvisionedDevices": []string{"00008020-000649AE3683002E"},
		"Entitlements":       map[string]interface{}{"get-task-allow": true},
	})
	writeZip(t, ipa, map[string][]byte{
		"Payload/Game.app/Info.plist": mustPlist(t, map[string]interface{}{
			"CFBundleIdentifier": "com.example.game", "CFBundleShortVersionString": "1.2.3", "CFBundleVersion": "45",
			"CFBundleDisplayName": "Game", "MinimumOSVersion": "15.0", "DTPlatformName": "iphoneos", "DTXcode": "1610",
			"UIDeviceFamily": []int{1, 2}, "NSCameraUsageDescription": "Scan codes",
			"CFBundleIcons": map[string]interface{}{"CFBundlePrimaryIcon": map[string]interface{}{"CFBundleIconFiles": []string{"AppIcon60x60"}}},
		}),
		"Payload/Game.app/embedded.mobileprovision":                           append(append([]byte("\x30\x80junk"), profile...), []byte("sig")...),
		"Payload/Game.app/AppIcon60x60@2x.png":                                []byte("small"),
		"Payload/Game.app/AppIcon60x60@3x.png":                                []byte("larger icon"),
		"Payload/Game.app/Frameworks/UnityFramework.framework/UnityFramework": []byte("bin"),
		"Payload/Game.app/Data/globalgamemanagers":                            []byte("\x00\x00\x00\x002022.3.20f1\x00"),
		"Payload/Game.app/Frameworks/UnityFramework.framework/Data/x":         []byte("x"),
		"Payload/Game.app/PlugIns/Widget.appex/Info.plist":                    mustPlist(t, map[string]interface{}{"CFBundleIdentifier": "com.example.game.widget"}),
	})
	a := NewAnalyser(ipa)
	if err := a.AnalyseFile(); err != nil {
		t.Fatal(err)
	}
	p := a.GetParameter()
	if p.Platform != "ios" || p.Identifier != "com.example.game" || p.Version != "45" || p.Name != "Game" {
		t.Fatalf("parameter = %+v", p)
	}
	info := p.Info
	if info.Version != "1.2.3" || info.Apple.MinOS != "15.0" || info.Apple.Xcode != "16.1" {
		t.Errorf("info = %+v apple = %+v", info, info.Apple)
	}
	if len(info.Apple.DeviceFamily) != 2 || info.Apple.DeviceFamily[1] != "iPad" {
		t.Errorf("families = %v", info.Apple.DeviceFamily)
	}
	if len(info.Apple.UsageDescriptions) != 1 || len(info.Apple.Extensions) != 1 {
		t.Errorf("usage %v extensions %v", info.Apple.UsageDescriptions, info.Apple.Extensions)
	}
	if info.Icon != "Payload/Game.app/AppIcon60x60@3x.png" {
		t.Errorf("icon = %s", info.Icon)
	}
	if info.Engine == nil || info.Engine.Name != "Unity" || info.Engine.Version != "2022.3.20f1" {
		t.Errorf("engine = %+v", info.Engine)
	}
	if pr := info.Apple.Profile; pr == nil || pr.Kind != "development" || pr.TeamID != "ABCDE12345" || len(pr.Devices) != 1 {
		t.Errorf("profile = %+v", pr)
	}
	if info.Contents == nil || info.Contents.Files != 8 {
		t.Errorf("contents = %+v", info.Contents)
	}
}

func TestAnalyseZippedMacApp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Game.zip")
	writeZip(t, path, map[string][]byte{
		"Game.app/Contents/Info.plist":   mustPlist(t, map[string]interface{}{"CFBundleIdentifier": "com.example.mac", "CFBundleVersion": "7", "CFBundleExecutable": "Game", "LSMinimumSystemVersion": "12.0"}),
		"Game.app/Contents/MacOS/Game":   []byte("bin"),
		"__MACOSX/Game.app/._Info.plist": []byte("x"),
	})
	a := NewAnalyser(path)
	if err := a.AnalyseFile(); err != nil {
		t.Fatal(err)
	}
	p := a.GetParameter()
	if p.Platform != "macos" || p.Name != "Game.app" || p.Version != "7" || p.App.Executable.Executable != "Contents/MacOS/Game" {
		t.Fatalf("parameter = %+v %+v", p, p.App.Executable)
	}
	if p.Info.Apple.MinOS != "12.0" {
		t.Errorf("min os = %s", p.Info.Apple.MinOS)
	}
}

// crush makes an Xcode style png: BGRA, premultiplied, raw deflate, CgBI
func crush(t *testing.T, img *image.NRGBA) []byte {
	var normal bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.NoCompression}).Encode(&normal, img); err != nil {
		t.Fatal(err)
	}
	chunks, err := pngChunks(normal.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	var idat bytes.Buffer
	var ihdr []byte
	for _, c := range chunks {
		if c.kind == "IDAT" {
			idat.Write(c.data)
		}
		if c.kind == "IHDR" {
			ihdr = c.data
		}
	}
	zr, err := zlib.NewReader(&idat)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(zr)
	w := img.Bounds().Dx()
	stride := 1 + w*4
	for y := 0; y < img.Bounds().Dy(); y++ {
		row := raw[y*stride+1 : (y+1)*stride]
		if raw[y*stride] != 0 {
			t.Fatalf("expected no filter")
		}
		for x := 0; x < len(row); x += 4 {
			a := int(row[x+3])
			r, g, b := int(row[x])*a/255, int(row[x+1])*a/255, int(row[x+2])*a/255
			row[x], row[x+1], row[x+2] = uint8(b), uint8(g), uint8(r)
		}
	}
	var deflated bytes.Buffer
	fw, _ := flate.NewWriter(&deflated, flate.BestCompression)
	fw.Write(raw)
	fw.Close()
	var out bytes.Buffer
	out.Write(pngSignature)
	writeChunk(&out, "CgBI", []byte{0x50, 0x00, 0x20, 0x06})
	writeChunk(&out, "IHDR", ihdr)
	writeChunk(&out, "IDAT", deflated.Bytes())
	writeChunk(&out, "IEND", nil)
	return out.Bytes()
}

func TestUncrushPNG(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 10, B: 0, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 0, G: 0, B: 200, A: 255})
	img.SetNRGBA(2, 1, color.NRGBA{R: 200, G: 100, B: 50, A: 128})
	fixed, err := uncrushPNG(crush(t, img))
	if err != nil {
		t.Fatal(err)
	}
	got, err := png.Decode(bytes.NewReader(fixed))
	if err != nil {
		t.Fatal(err)
	}
	near := func(a, b uint8) bool { return int(a)-int(b) <= 2 && int(b)-int(a) <= 2 }
	for _, p := range []image.Point{{0, 0}, {1, 0}, {2, 1}} {
		want := img.NRGBAAt(p.X, p.Y)
		have := color.NRGBAModel.Convert(got.At(p.X, p.Y)).(color.NRGBA)
		if !near(want.R, have.R) || !near(want.G, have.G) || !near(want.B, have.B) || want.A != have.A {
			t.Errorf("pixel %v = %v, want %v", p, have, want)
		}
	}
}

func TestPNGFromICNS(t *testing.T) {
	small := append(append([]byte{}, pngSignature...), 1)
	big := append(append([]byte{}, pngSignature...), 1, 2, 3)
	var icns bytes.Buffer
	icns.WriteString("icns\x00\x00\x00\x00")
	for _, e := range []struct {
		kind string
		data []byte
	}{{"ic07", small}, {"is32", []byte("rawdata")}, {"ic10", big}} {
		icns.WriteString(e.kind)
		n := len(e.data) + 8
		icns.Write([]byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)})
		icns.Write(e.data)
	}
	if got := pngFromICNS(icns.Bytes()); !bytes.Equal(got, big) {
		t.Errorf("got %v", got)
	}
}
