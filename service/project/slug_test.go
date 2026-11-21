package project

import "testing"

func TestSlug(t *testing.T) {
	tests := map[string]string{
		"Heroes of History": "heroes-of-history",
		"  R&D / Tests? ":   "r-d-tests",
		"Über Spiel 2":      "ber-spiel-2",
		"---":               "project",
		"日本":                "project",
		"a  --  b":          "a-b",
		"Trailing!!!":       "trailing",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa b": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	for name, want := range tests {
		if got := slug(name); got != want {
			t.Errorf("%q: %q, want %q", name, got, want)
		}
	}
}
