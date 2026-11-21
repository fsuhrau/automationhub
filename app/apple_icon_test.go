package app

import (
	"reflect"
	"testing"
)

func TestIconNames(t *testing.T) {
	content := plistDict{
		"CFBundleIcons":      plistDict{"CFBundlePrimaryIcon": plistDict{"CFBundleIconFiles": []interface{}{"A", "B"}, "CFBundleIconName": "AppIcon"}},
		"CFBundleIcons~ipad": plistDict{"CFBundlePrimaryIcon": plistDict{"CFBundleIconFiles": []interface{}{"C"}}},
		"CFBundleIconFiles":  []interface{}{"D"},
		"CFBundleIconFile":   "E.icns",
	}
	want := []string{"A", "B", "AppIcon", "C", "D", "E.icns"}
	if got := iconNames(content); !reflect.DeepEqual(got, want) {
		t.Errorf("names = %v, want %v", got, want)
	}
}

func TestIconNamesOfBrokenDicts(t *testing.T) {
	for name, icons := range map[string]interface{}{
		"missing": nil, "not a dict": "x", "no primary icon": plistDict{}, "primary is not a dict": plistDict{"CFBundlePrimaryIcon": "x"},
	} {
		if got := primaryIconNames(icons); got != nil {
			t.Errorf("%s: names = %v", name, got)
		}
	}
	if got := iconNames(plistDict{}); got != nil {
		t.Errorf("names = %v", got)
	}
}

func TestAppleBundleIcon(t *testing.T) {
	names := plistDict{"CFBundleIconFiles": []interface{}{"AppIcon60x60"}}
	tests := []struct {
		name    string
		bundle  appleBundle
		files   map[string][]byte
		content plistDict
		want    string
	}{
		{"largest scale", appleBundle{root: "P/G.app/"}, map[string][]byte{
			"P/G.app/AppIcon60x60@2x.png": []byte("ab"), "P/G.app/AppIcon60x60@3x.png": []byte("abc"), "P/G.app/AppIcon60x60~ipad.png": []byte("a"),
			"P/G.app/Other.png": []byte("abcdefgh"), "P/G.app/AppIcon60x60.jpg": []byte("abcdefgh"),
			"P/G.app/sub/AppIcon60x60.png": []byte("abcdefgh"), "X/AppIcon60x60.png": []byte("abcdefgh"),
		}, names, "P/G.app/AppIcon60x60@3x.png"},
		{"mac icns in Resources", appleBundle{root: "G.app/Contents/", mac: true}, map[string][]byte{
			"G.app/Contents/Resources/Icon.ICNS": []byte("x"), "G.app/Contents/Icon.icns": []byte("xxxx"),
		}, plistDict{"CFBundleIconFile": "Icon.icns"}, "G.app/Contents/Resources/Icon.ICNS"},
		{"no names in the plist", appleBundle{root: "G.app/"}, map[string][]byte{"G.app/Icon.png": []byte("x")}, plistDict{}, ""},
		{"no match", appleBundle{root: "G.app/"}, map[string][]byte{"G.app/Nope.png": []byte("x")}, names, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, tt.bundle.files = openZip(t, tt.files)
			if got := tt.bundle.icon(tt.content); got != tt.want {
				t.Errorf("icon = %q, want %q", got, tt.want)
			}
		})
	}
}
