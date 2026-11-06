package androiddevice

import "testing"

func TestParseWmSize(t *testing.T) {
	cases := []struct {
		out  string
		w, h int
	}{
		{"Physical size: 1080x2400\n", 1080, 2400},
		{"Physical size: 1440x3200\nOverride size: 1080x2400\n", 1080, 2400},
		{"error: no devices", 0, 0},
	}
	for _, c := range cases {
		if w, h := parseWmSize(c.out); w != c.w || h != c.h {
			t.Errorf("parseWmSize(%q) = %dx%d, want %dx%d", c.out, w, h, c.w, c.h)
		}
	}
}

func TestKeyCode(t *testing.T) {
	if c, err := keyCode("BACK"); err != nil || c != KEYCODE_BACK {
		t.Fatalf("back: %d %v", c, err)
	}
	if c, err := keyCode("66"); err != nil || c != 66 {
		t.Fatalf("numeric: %d %v", c, err)
	}
	if _, err := keyCode("nope"); err == nil {
		t.Fatal("expected an error for an unknown key")
	}
}
