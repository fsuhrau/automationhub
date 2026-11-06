package device

import (
	"errors"
	"fmt"
	"testing"
)

func TestControlErrorRestoresUnsupported(t *testing.T) {
	wrapped := fmt.Errorf("live video needs WebDriverAgent: %w", ErrUnsupported)
	for _, err := range []error{ErrUnsupported, wrapped} {
		got := ControlError(err.Error())
		if !errors.Is(got, ErrUnsupported) || got.Error() != err.Error() {
			t.Errorf("ControlError(%q) = %v", err.Error(), got)
		}
	}
	if errors.Is(ControlError("boom"), ErrUnsupported) {
		t.Error("a plain error must not be unsupported")
	}
}
