package app

import "testing"

func TestAPKIcon(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		icons map[int]string
		want  string
	}{
		{"png of the biggest density that exists", []string{"res/a.png", "res/b.png"},
			map[int]string{160: "res/a.png", 320: "res/b.png", 640: "res/missing.png"}, "res/b.png"},
		{"webp is raster", []string{"res/a.webp"}, map[int]string{160: "res/a.webp"}, "res/a.webp"},
		{"adaptive icon uses the biggest density png", []string{"res/mipmap-mdpi/ic.png", "res/mipmap-xxhdpi/ic.png", "res/mipmap-anydpi-v26/ic.xml"},
			map[int]string{640: "res/mipmap-anydpi-v26/ic.xml"}, "res/mipmap-xxhdpi/ic.png"},
		{"drawable webp without density", []string{"res/drawable-v24/ic.webp"}, map[int]string{0: "res/drawable-v24/ic.xml"}, "res/drawable-v24/ic.webp"},
		{"other names and folders are ignored", []string{"res/mipmap-hdpi/other.png", "res/raw/ic.png", "res/mipmap-hdpi/ic.xml", "res/mipmap-hdpi/.png", "res/mipmap-hdpi/ic"},
			map[int]string{160: "res/mipmap-anydpi/ic.xml"}, ""},
		{"no icons", []string{"res/mipmap-hdpi/ic.png"}, map[int]string{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := apkIcon(zipFiles(t, tt.files...), tt.icons); got != tt.want {
				t.Errorf("apkIcon = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDensityRank(t *testing.T) {
	for path, want := range map[string]int{
		"res/mipmap-xxxhdpi/a.png": 10, "res/mipmap-xxhdpi/a.png": 9, "res/mipmap-xhdpi/a.png": 8,
		"res/mipmap-hdpi/a.png": 7, "res/mipmap-mdpi/a.png": 6, "res/mipmap-ldpi/a.png": 5, "res/drawable-v24/a.png": 0,
	} {
		if got := densityRank(path); got != want {
			t.Errorf("densityRank(%s) = %d, want %d", path, got, want)
		}
	}
}

func TestIconPathsByDensityBiggestFirst(t *testing.T) {
	got := iconPathsByDensity(map[int]string{160: "a", 640: "c", 320: "b"})
	if len(got) != 3 || got[0] != "c" || got[1] != "b" || got[2] != "a" {
		t.Errorf("paths = %v", got)
	}
}
