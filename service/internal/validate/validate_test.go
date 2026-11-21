package validate_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/internal/validate"
)

func TestTrimRequired(t *testing.T) {
	tests := []struct {
		name, in, want string
		invalid        bool
	}{
		{"plain", "game", "game", false},
		{"trimmed", "  game \n", "game", false},
		{"empty", "", "", true},
		{"blank", " \t ", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validate.TrimRequired(tc.in, "needs a name")
			if got != tc.want || (apperr.KindOf(err) == apperr.KindInvalid) != tc.invalid {
				t.Fatalf("got %q, %v", got, err)
			}
			if tc.invalid && err.Error() != "needs a name" {
				t.Fatalf("message %q", err)
			}
		})
	}
}
