package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultPathsLiveBelowTheHub(t *testing.T) {
	got := DefaultPaths("/home/u")

	want := Paths{
		AppStore: filepath.Join("/home/u", ".automationhub", "store"),
		Bundles:  filepath.Join("/home/u", ".automationhub", "bundles"),
		TestData: filepath.Join("/home/u", ".automationhub", "data"),
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestWithDefaultsKeepsConfiguredPaths(t *testing.T) {
	defaults := Paths{AppStore: "a", Bundles: "b", TestData: "c"}
	tests := []struct {
		name string
		in   Paths
		want Paths
	}{
		{"empty", Paths{}, defaults},
		{"partly set", Paths{Bundles: "x"}, Paths{AppStore: "a", Bundles: "x", TestData: "c"}},
		{"all set", Paths{AppStore: "1", Bundles: "2", TestData: "3"}, Paths{AppStore: "1", Bundles: "2", TestData: "3"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.in.WithDefaults(defaults); got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestAllListsEveryFolder(t *testing.T) {
	got := Paths{AppStore: "a", Bundles: "b", TestData: "c"}.All()

	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("got %v", got)
	}
}
