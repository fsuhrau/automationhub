package flowtest

import (
	"fmt"
	"testing"
)

// recorder is a testing.TB that records Fatalf instead of stopping.
type recorder struct {
	testing.TB
	failed string
}

func (r *recorder) Helper()                   {}
func (r *recorder) Fatalf(f string, a ...any) { r.failed = fmt.Sprintf(f, a...) }

func TestCheck(t *testing.T) {
	r := &recorder{}
	check(r, true, "never")
	if r.failed != "" {
		t.Fatalf("a passing check failed: %s", r.failed)
	}
	check(r, false, "bad %d", 7)
	if r.failed != "bad 7" {
		t.Fatalf("a failing check said %q", r.failed)
	}
}
