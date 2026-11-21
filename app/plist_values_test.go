package app

import (
	"reflect"
	"testing"
)

func TestPlistString(t *testing.T) {
	m := plistDict{"s": "text", "u": uint64(7), "i": int64(-3), "f": 1.5, "b": true, "list": []interface{}{"x"}}
	for key, want := range map[string]string{"s": "text", "u": "7", "i": "-3", "f": "1.5", "b": "true", "list": "", "missing": ""} {
		if got := plistString(m, key); got != want {
			t.Errorf("plistString(%s) = %q, want %q", key, got, want)
		}
	}
}

func TestPlistStrings(t *testing.T) {
	m := plistDict{
		"list":    []interface{}{"a", uint64(2), int64(3), 1.5, true},
		"dict":    plistDict{"b": true, "a": true, "off": false, "odd": "yes"},
		"single":  "one",
		"unknown": 5,
	}
	tests := map[string][]string{
		"list":    {"a", "2", "3"},
		"dict":    {"a", "b", "odd"},
		"single":  {"one"},
		"unknown": nil,
		"missing": nil,
	}
	for key, want := range tests {
		if got := plistStrings(m, key); !reflect.DeepEqual(got, want) {
			t.Errorf("plistStrings(%s) = %q, want %q", key, got, want)
		}
	}
}
