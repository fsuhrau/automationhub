package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
)

func TestConstructorsSetKindAndMessage(t *testing.T) {
	tests := []struct {
		name string
		make func(string, ...any) error
		want apperr.Kind
	}{
		{"not found", apperr.NotFound, apperr.KindNotFound},
		{"invalid", apperr.Invalid, apperr.KindInvalid},
		{"conflict", apperr.Conflict, apperr.KindConflict},
		{"busy", apperr.Busy, apperr.KindBusy},
		{"precondition required", apperr.PreconditionRequired, apperr.KindPreconditionRequired},
		{"unsupported", apperr.Unsupported, apperr.KindUnsupported},
		{"timeout", apperr.Timeout, apperr.KindTimeout},
		{"forbidden", apperr.Forbidden, apperr.KindForbidden},
		{"internal", apperr.Internal, apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.make("test run %d", 7)
			if got := apperr.KindOf(err); got != tc.want {
				t.Fatalf("kind = %v, want %v", got, tc.want)
			}
			if err.Error() != "test run 7" {
				t.Fatalf("message = %q", err.Error())
			}
		})
	}
}

func TestConstructorKeepsWrappedCause(t *testing.T) {
	cause := errors.New("disk")
	err := apperr.Internal("save: %w", cause)

	if !errors.Is(err, cause) {
		t.Fatal("cause is not reachable")
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("record not found")
	err := apperr.Wrap(apperr.KindNotFound, cause, "load run")

	if err.Error() != "load run: record not found" {
		t.Fatalf("message = %q", err.Error())
	}
	if !errors.Is(err, cause) {
		t.Fatal("cause is not reachable")
	}
	if apperr.KindOf(err) != apperr.KindNotFound {
		t.Fatalf("kind = %v", apperr.KindOf(err))
	}
}

func TestWrapNilIsNil(t *testing.T) {
	if err := apperr.Wrap(apperr.KindInvalid, nil, "x"); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestKindOf(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want apperr.Kind
	}{
		{"nil", nil, apperr.KindNone},
		{"plain error", errors.New("boom"), apperr.KindInternal},
		{"wrapped by fmt", fmt.Errorf("outer: %w", apperr.Busy("device")), apperr.KindBusy},
		{"outermost kind wins", apperr.Wrap(apperr.KindConflict, apperr.NotFound("x"), "y"), apperr.KindConflict},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := apperr.KindOf(tc.err); got != tc.want {
				t.Fatalf("kind = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestErrorExposesItsKind(t *testing.T) {
	var e *apperr.Error
	if !errors.As(apperr.Timeout("slow"), &e) || e.Kind() != apperr.KindTimeout {
		t.Fatalf("kind = %v", e)
	}
}
