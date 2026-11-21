package cmd

import (
	"strings"
	"testing"
)

func TestTestCommandListsTheTests(t *testing.T) {
	hub := newFakeHub(t)

	out, err := execute(t, "--token", "secret", "test", hub.URL, "demo", "7")

	if err != nil {
		t.Fatal(err)
	}
	if want := "ID: 3 Name: smoke\nID: 4 Name: login\n"; out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
	if got := hub.seen(); len(got) != 1 || !strings.HasPrefix(got[0], "GET "+appPrefix+"/tests") || hub.tokens[0] != "secret" {
		t.Errorf("requests = %v, tokens = %v", got, hub.tokens)
	}
}

func TestTestCommandFailures(t *testing.T) {
	hub := newFakeHub(t)
	hub.failOn(appPrefix+"/tests", 500)
	tests := map[string]struct {
		args    []string
		wantErr string
	}{
		"missing parameters": {[]string{"test", hub.URL}, "missing parameter"},
		"bad app id":         {[]string{"test", hub.URL, "demo", "x"}, "invalid appID"},
		"hub error":          {[]string{"test", hub.URL, "demo", "7"}, "hub says no"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := execute(t, tc.args...); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("err = %v, want %q", err, tc.wantErr)
			}
		})
	}
}
