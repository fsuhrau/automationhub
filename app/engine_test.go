package app

import (
	"reflect"
	"slices"
	"testing"
)

const managersFile = "Data/globalgamemanagers"

func TestDetectEngine(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  *Engine
	}{
		{"unity il2cpp", []string{"lib/x/libunity.so", "lib/x/libil2cpp.so", "assets/bin/Data/Managed/mscorlib.dll"}, &Engine{Name: "Unity", Backend: "IL2CPP"}},
		{"unity il2cpp metadata", []string{"lib/x/libunity.so", "Managed/Metadata/global-metadata.dat"}, &Engine{Name: "Unity", Backend: "IL2CPP"}},
		{"unity mono runtime", []string{"lib/x/libunity.so", "lib/x/libmonobdwgc-2.0.so"}, &Engine{Name: "Unity", Backend: "Mono"}},
		{"unity mono sgen", []string{"lib/x/libunity.so", "lib/x/libmonosgen-2.0.so"}, &Engine{Name: "Unity", Backend: "Mono"}},
		{"unity mono corlib", []string{"UnityPlayer.dll", "Game_Data/Managed/mscorlib.dll"}, &Engine{Name: "Unity", Backend: "Mono"}},
		{"unity framework", []string{"Frameworks/UnityFramework.framework/UnityFramework"}, &Engine{Name: "Unity"}},
		{"unity dylib", []string{"UnityPlayer.dylib"}, &Engine{Name: "Unity"}},
		{"unity version", []string{"lib/x/libunity.so", managersFile}, &Engine{Name: "Unity", Version: "2022.3.20f1"}},
		{"unity by its managers file", []string{managersFile}, &Engine{Name: "Unity", Version: "2022.3.20f1"}},
		{"unity wins over another engine", []string{"lib/x/libflutter.so", "lib/x/libunity.so"}, &Engine{Name: "Unity"}},
		{"unreal", []string{"lib/x/libUE4.so"}, &Engine{Name: "Unreal Engine"}},
		{"unreal 5", []string{"lib/x/libUnreal.so"}, &Engine{Name: "Unreal Engine"}},
		{"flutter", []string{"lib/x/libflutter.so"}, &Engine{Name: "Flutter"}},
		{"flutter framework", []string{"Frameworks/Flutter.framework/Flutter"}, &Engine{Name: "Flutter"}},
		{"react native", []string{"assets/index.android.bundle"}, &Engine{Name: "React Native"}},
		{"godot", []string{"lib/x/libgodot_android.so"}, &Engine{Name: "Godot"}},
		{"cocos", []string{"lib/x/libcocos2dcpp.so"}, &Engine{Name: "Cocos"}},
		{"nothing known", []string{"classes.dex"}, nil},
		{"nothing at all", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string][]byte{}
			for _, name := range tt.files {
				files[name] = []byte(name)
			}
			if slices.Contains(tt.files, managersFile) {
				files[managersFile] = []byte("\x00\x00\x002022.3.20f1\x00")
			}
			_, entries := openZip(t, files)
			if got := detectEngine(entries, ""); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("engine = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDetectEngineLooksOnlyBelowTheRoot(t *testing.T) {
	entries := zipFiles(t, "Payload/G.app/Frameworks/UnityFramework.framework/U", "Other.app/libflutter.so")
	if got := detectEngine(entries, "Payload/G.app/"); got == nil || got.Name != "Unity" {
		t.Errorf("engine = %+v", got)
	}
	if got := detectEngine(entries, "Missing/"); got != nil {
		t.Errorf("engine = %+v", got)
	}
}

func TestDetectEngineTakesTheVersionOfTheShortestManagersFile(t *testing.T) {
	_, entries := openZip(t, map[string][]byte{
		"a/b/c/globalgamemanagers": []byte("2021.1.1f1"), "a/globalgamemanagers": []byte("2022.3.20f1"),
	})
	if got := detectEngine(entries, ""); got == nil || got.Version != "2022.3.20f1" {
		t.Errorf("engine = %+v", got)
	}
}

func TestUnityVersionOfAnUnreadableFile(t *testing.T) {
	_, entries := openZip(t, nil, "Data/globalgamemanagers")
	if got := unityVersion(entries[0]); got != "" {
		t.Errorf("version = %q", got)
	}
}
