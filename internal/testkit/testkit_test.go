package testkit

import (
	"errors"
	"testing"
)

// fatalTB turns Fatalf into a panic so error paths of helpers can be tested.
type fatalTB struct {
	testing.TB
	msg  string
	name string
}

func (f *fatalTB) Name() string {
	if f.name != "" {
		return f.name
	}
	return f.TB.Name()
}

func (f *fatalTB) Fatalf(format string, _ ...any) { f.msg = format; panic(f) }
func (f *fatalTB) Helper()                        {}

// expectFatal runs fn and reports whether it ended in Fatalf.
func expectFatal(t *testing.T, name string, fn func(tb *fatalTB)) {
	t.Helper()
	tb := &fatalTB{TB: t, name: name}
	defer func() {
		if recover() != tb {
			t.Fatal("expected Fatalf")
		}
	}()
	fn(tb)
}

var errBoom = errors.New("boom")
