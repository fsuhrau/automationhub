package app

import "testing"

func TestAppleRoot(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		root    string
		mac     bool
		wantErr bool
	}{
		{"ipa", []string{"Payload/Game.app/Info.plist", "Payload/Game.app/x"}, "Payload/Game.app/", false, false},
		{"zipped mac app", []string{"Game.app/Contents/Info.plist"}, "Game.app/Contents/", true, false},
		{"zipped ios app", []string{"Game.app/Info.plist"}, "Game.app/", false, false},
		{"shortest path wins over plug-ins", []string{"Payload/Game.app/PlugIns/W.appex/Info.plist", "Payload/Game.app/Info.plist"}, "Payload/Game.app/", false, false},
		{"resource fork copies are skipped", []string{"__MACOSX/Game.app/Info.plist"}, "", false, true},
		{"no app", []string{"readme.txt"}, "", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, mac, err := appleRoot(zipFiles(t, tt.files...))
			if (err != nil) != tt.wantErr || root != tt.root || mac != tt.mac {
				t.Errorf("appleRoot = %q, %v, %v", root, mac, err)
			}
		})
	}
}
