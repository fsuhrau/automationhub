package androiddevice

import "testing"

func TestEscapeInputText(t *testing.T) {
	cases := map[string]string{
		"hello":         "hello",
		"hello world":   "hello%sworld",
		"a'b\"c":        `a\'b\"c`,
		"$(rm -rf /)":   `\$\(rm%s-rf%s/\)`,
		"50% off":       `50\%%soff`,
		"line1\nline2":  "line1line2",
		"user@host.com": "user@host.com",
	}
	for in, want := range cases {
		if got := escapeInputText(in); got != want {
			t.Errorf("escapeInputText(%q) = %q, want %q", in, got, want)
		}
	}
}
