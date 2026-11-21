package middleware

import (
	"errors"
	"io"
	"log"
	"strings"
	"testing"
)

func newTestLogger(w io.Writer) *log.Logger { return log.New(w, logPrefix, 0) }

func TestStackNamesThisFunction(t *testing.T) {
	got := string(stack(0))
	if !strings.Contains(got, "TestStackNamesThisFunction") || !strings.Contains(got, "stack_test.go") {
		t.Errorf("stack %q", got)
	}
}

func TestStackWithoutReadableSourceStillListsFrames(t *testing.T) {
	got := string(stackFrom(0, func(string) ([]byte, error) { return nil, errors.New("gone") }))
	if !strings.Contains(got, "stack_test.go") || strings.Contains(got, "\t") {
		t.Errorf("stack %q", got)
	}
}

func TestSource(t *testing.T) {
	lines := [][]byte{[]byte("  first "), []byte("second")}
	cases := []struct {
		n    int
		want string
	}{{1, "first"}, {2, "second"}, {0, dunno}, {3, dunno}}
	for _, c := range cases {
		if got := string(source(lines, c.n)); got != c.want {
			t.Errorf("source(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestShortName(t *testing.T) {
	cases := map[string]string{
		"runtime/debug.*T·ptrmethod":   "*T.ptrmethod",
		"code.google.com/p/x/pkg.Func": "Func",
		"main.main":                    "main",
		"plain":                        "plain",
	}
	for in, want := range cases {
		if got := shortName(in); got != want {
			t.Errorf("shortName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFunctionOfUnknownPC(t *testing.T) {
	if got := string(function(0)); got != dunno {
		t.Errorf("function(0) = %q, want %q", got, dunno)
	}
}
