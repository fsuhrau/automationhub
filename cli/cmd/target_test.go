package cmd

import (
	"strings"
	"testing"
)

func TestParseTarget(t *testing.T) {
	tests := map[string]struct {
		args    []string
		want    target
		wantErr string
	}{
		"complete":            {args: []string{"http://hub", "demo", "7"}, want: target{"http://hub", "demo", 7}},
		"extra arguments":     {args: []string{"http://hub", "demo", "7", "smoke"}, want: target{"http://hub", "demo", 7}},
		"missing":             {args: []string{"http://hub", "demo"}, wantErr: "missing parameter"},
		"no arguments":        {wantErr: "missing parameter"},
		"app is not a number": {args: []string{"http://hub", "demo", "seven"}, wantErr: `invalid appID "seven"`},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseTarget(tc.args)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("err = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Errorf("got %+v, %v, want %+v", got, err, tc.want)
			}
		})
	}
}
