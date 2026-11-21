package app

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"path/filepath"
	"testing"
)

func icnsOf(png []byte) []byte {
	var icns bytes.Buffer
	icns.WriteString("icns\x00\x00\x00\x00ic10")
	n := len(png) + 8
	icns.Write([]byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)})
	icns.Write(png)
	return icns.Bytes()
}

func TestIcon(t *testing.T) {
	plain := append(append([]byte{}, pngSignature...), 1, 2, 3)
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 9, A: 200})
	crushed := crush(t, img)
	badCrushed := corruptIDAT(crushed)
	path, _ := openZip(t, map[string][]byte{
		"a.png": plain, "crushed.png": crushed, "bad.png": badCrushed, "text.png": []byte("not a png"),
		"a.webp": []byte("webp"), "a.icns": icnsOf(plain), "empty.icns": []byte("icns"),
	}, "broken.png")

	tests := []struct {
		name     string
		app      string
		icon     string
		wantType string
		wantErr  error
		anyErr   bool
	}{
		{"png as it is", path, "a.png", "image/png", nil, false},
		{"crushed png is fixed", path, "crushed.png", "image/png", nil, false},
		{"webp", path, "a.webp", "image/webp", nil, false},
		{"icns", path, "a.icns", "image/png", nil, false},
		{"icns without png", path, "empty.icns", "", ErrNoIcon, false},
		{"not a png", path, "text.png", "", ErrNoIcon, false},
		{"no icon path", path, "", "", ErrNoIcon, false},
		{"icon not in the archive", path, "missing.png", "", ErrNoIcon, false},
		{"crushed png is broken", path, "bad.png", "", nil, true},
		{"entry cannot be opened", path, "broken.png", "", nil, true},
		{"archive is missing", filepath.Join(t.TempDir(), "none.zip"), "a.png", "", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, contentType, err := Icon(tt.app, tt.icon)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) || tt.anyErr && err == nil || tt.wantErr == nil && !tt.anyErr && err != nil {
				t.Fatalf("err = %v", err)
			}
			if contentType != tt.wantType || (err == nil) != (data != nil) {
				t.Errorf("content type = %q, data = %v", contentType, data)
			}
		})
	}
}

func TestIconReturnsTheStoredBytes(t *testing.T) {
	plain := append(append([]byte{}, pngSignature...), 1, 2, 3)
	path, _ := openZip(t, map[string][]byte{"a.png": plain, "a.icns": icnsOf(plain)})
	for _, name := range []string{"a.png", "a.icns"} {
		if data, _, err := Icon(path, name); err != nil || !bytes.Equal(data, plain) {
			t.Errorf("%s: data = %v, err = %v", name, data, err)
		}
	}
}

func TestPNGFromICNSIgnoresBrokenFiles(t *testing.T) {
	for name, data := range map[string][]byte{
		"too short": []byte("icns"), "wrong magic": []byte("nope\x00\x00\x00\x00"),
		"entry too small": []byte("icns\x00\x00\x00\x00ic10\x00\x00\x00\x04"),
		"entry too long":  []byte("icns\x00\x00\x00\x00ic10\x00\x00\x01\x00"),
	} {
		if got := pngFromICNS(data); got != nil {
			t.Errorf("%s: png = %v", name, got)
		}
	}
}
