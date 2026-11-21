package api

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestStatusForError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"not found", fmt.Errorf("x: %w", ErrNotFound), http.StatusNotFound},
		{"bundle exists", ErrBundleExists, http.StatusConflict},
		{"no devices", ErrNoDevices, http.StatusConflict},
		{"invalid request", badRequest("bad"), http.StatusBadRequest},
		{"anything else", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := statusForError(tc.err); got != tc.want {
				t.Fatalf("status %d, want %d", got, tc.want)
			}
		})
	}
}
