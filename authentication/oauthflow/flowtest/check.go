package flowtest

import "testing"

// check stops the test with the message when ok is false. Every assertion of
// the suite goes through it, so the suite has a single failure branch.
//
//nolint:revive // the signature mirrors testing.TB.Fatalf
func check(t testing.TB, ok bool, format string, args ...any) {
	t.Helper()
	if !ok {
		t.Fatalf(format, args...)
	}
}
