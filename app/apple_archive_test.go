package app

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func infoPlist(t *testing.T, extra map[string]interface{}) []byte {
	t.Helper()
	m := map[string]interface{}{"CFBundleIdentifier": "com.x", "CFBundleExecutable": "Game"}
	for k, v := range extra {
		m[k] = v
	}
	return mustPlist(t, m)
}

func TestReadAppleArchiveIOS(t *testing.T) {
	cert := testCertificate(t, "Dev")
	profile := signedProfile(t, provisioningPlist{Name: "P", DeveloperCertificates: [][]byte{cert.Raw}})
	_, files := openZip(t, map[string][]byte{
		"Payload/G.app/Info.plist":                        infoPlist(t, map[string]interface{}{"CFBundleIconFiles": []string{"Icon"}}),
		"Payload/G.app/Icon@2x.png":                       []byte("png"),
		"Payload/G.app/embedded.mobileprovision":          profile,
		"Payload/G.app/Frameworks/A.framework/A":          []byte("a"),
		"Payload/G.app/Frameworks/A.framework/Info.plist": []byte("a"),
		"Payload/G.app/PlugIns/W.appex/W":                 []byte("w"),
		"Payload/G.app/PlugIns/T.xctest/T":                []byte("t"),
		"Payload/G.app/PlugIns/readme.txt":                []byte("r"),
		"Payload/G.app/Extensions/E.appex/E":              []byte("e"),
		"Payload/Other.txt":                               []byte("o"),
	})
	info := &BundleInfo{}
	root, err := readAppleArchive(files, info)
	if err != nil {
		t.Fatal(err)
	}
	if root != "Payload/G.app/" || info.Identifier != "com.x" || info.Icon != "Payload/G.app/Icon@2x.png" {
		t.Errorf("root = %s info = %+v", root, info)
	}
	if info.Apple.Executable != "Game" || info.Apple.Profile == nil || info.Apple.Profile.Name != "P" {
		t.Errorf("apple = %+v", info.Apple)
	}
	if !reflect.DeepEqual(info.Apple.Frameworks, []string{"A.framework"}) {
		t.Errorf("frameworks = %v", info.Apple.Frameworks)
	}
	if got := info.Apple.Extensions; len(got) != 3 {
		t.Errorf("extensions = %v", got)
	}
	if info.Signing == nil || !reflect.DeepEqual(info.Signing.Schemes, []string{"profile"}) || len(info.Signing.Certificates) != 1 {
		t.Errorf("signing = %+v", info.Signing)
	}
}

func TestReadAppleArchiveMac(t *testing.T) {
	_, files := openZip(t, map[string][]byte{
		"G.app/Contents/Info.plist":                infoPlist(t, map[string]interface{}{"CFBundleIconFile": "Icon"}),
		"G.app/Contents/Resources/Icon.icns":       []byte("icns"),
		"G.app/Contents/embedded.provisionprofile": signedProfile(t, provisioningPlist{Name: "Mac"}),
	})
	info := &BundleInfo{}
	if _, err := readAppleArchive(files, info); err != nil {
		t.Fatal(err)
	}
	if info.Apple.Platform != "macosx" || info.Apple.Executable != "Contents/MacOS/Game" || info.Icon != "G.app/Contents/Resources/Icon.icns" {
		t.Errorf("info = %+v apple = %+v", info, info.Apple)
	}
	if info.Apple.Profile == nil || info.Apple.Profile.Name != "Mac" || info.Signing != nil {
		t.Errorf("profile = %+v signing = %+v", info.Apple.Profile, info.Signing)
	}
}

func TestReadAppleArchiveKeepsMacPlatformFromPlist(t *testing.T) {
	_, files := openZip(t, map[string][]byte{"G.app/Contents/Info.plist": infoPlist(t, map[string]interface{}{"DTPlatformName": "macosx14", "CFBundleExecutable": ""})})
	info := &BundleInfo{}
	if _, err := readAppleArchive(files, info); err != nil {
		t.Fatal(err)
	}
	if info.Apple.Platform != "macosx14" || info.Apple.Executable != "" {
		t.Errorf("apple = %+v", info.Apple)
	}
}

func TestReadAppleArchiveWarnsAboutUnreadableProfile(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string][]byte
		broken  []string
		warning string
	}{
		{"no plist in the profile", map[string][]byte{"G.app/embedded.mobileprovision": []byte("junk")}, nil, "no plist in the provisioning profile"},
		{"entry cannot be opened", nil, []string{"G.app/embedded.mobileprovision"}, "unsupported compression"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.files == nil {
				tt.files = map[string][]byte{}
			}
			tt.files["G.app/Info.plist"] = infoPlist(t, nil)
			_, files := openZip(t, tt.files, tt.broken...)
			info := &BundleInfo{}
			if _, err := readAppleArchive(files, info); err != nil {
				t.Fatal(err)
			}
			if len(info.Warnings) != 1 || !strings.Contains(info.Warnings[0], tt.warning) || info.Apple.Profile != nil {
				t.Errorf("warnings = %v profile = %v", info.Warnings, info.Apple.Profile)
			}
		})
	}
}

func TestReadAppleArchiveErrors(t *testing.T) {
	tests := []struct {
		name   string
		files  map[string][]byte
		broken []string
	}{
		{"no app", map[string][]byte{"readme.txt": []byte("x")}, nil},
		{"info plist cannot be opened", nil, []string{"G.app/Info.plist"}},
		{"info plist is garbage", map[string][]byte{"G.app/Info.plist": []byte("garbage")}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, files := openZip(t, tt.files, tt.broken...)
			root, err := readAppleArchive(files, &BundleInfo{})
			if err == nil || root != "" {
				t.Errorf("root = %q, err = %v", root, err)
			}
		})
	}
}

func TestComponentOf(t *testing.T) {
	tests := []struct {
		rel, dir, name string
		ok             bool
	}{
		{"PlugIns/W.appex/W", "PlugIns/", "W.appex", true},
		{"Frameworks/A.framework", "Frameworks/", "A.framework", true},
		{"Extensions/E.appex/x", "Extensions/", "E.appex", true},
		{"PlugIns/", "PlugIns/", "", false},
		{"Info.plist", "", "", false},
	}
	for _, tt := range tests {
		dir, name, ok := componentOf(tt.rel)
		if dir != tt.dir || name != tt.name || ok != tt.ok {
			t.Errorf("componentOf(%s) = %q, %q, %v", tt.rel, dir, name, ok)
		}
	}
}

func TestReadAppleFolder(t *testing.T) {
	ios := t.TempDir()
	writeFile(t, filepath.Join(ios, "Info.plist"), infoPlist(t, nil))
	mac := t.TempDir()
	writeFile(t, filepath.Join(mac, "Contents", "Info.plist"), infoPlist(t, nil))
	garbage := t.TempDir()
	writeFile(t, filepath.Join(garbage, "Info.plist"), []byte("garbage"))

	tests := []struct {
		name     string
		path     string
		wantMac  bool
		wantExec string
		wantErr  bool
	}{
		{"ios", ios, false, "Game", false},
		{"mac", mac, true, "Contents/MacOS/Game", false},
		{"missing", filepath.Join(t.TempDir(), "none"), true, "", true},
		{"garbage", garbage, false, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &BundleInfo{}
			gotMac, err := readAppleFolder(tt.path, info)
			if (err != nil) != tt.wantErr || gotMac != tt.wantMac {
				t.Fatalf("mac = %v, err = %v", gotMac, err)
			}
			if !tt.wantErr && info.Apple.Executable != tt.wantExec {
				t.Errorf("executable = %s", info.Apple.Executable)
			}
		})
	}
}
