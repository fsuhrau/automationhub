package apitest

import (
	"errors"
	"fmt"
	"testing"
)

type recordingTB struct {
	testing.TB
	fatal string
}

func (r *recordingTB) Helper() {}

func (r *recordingTB) Fatalf(format string, args ...any) { r.fatal = fmt.Sprintf(format, args...) }

func TestFailOnStopsOnlyOnAnError(t *testing.T) {
	ok, bad := &recordingTB{}, &recordingTB{}

	failOn(ok, nil)
	failOn(bad, errors.New("boom"))

	if ok.fatal != "" || bad.fatal != "apitest: register routes: boom" {
		t.Fatalf("ok %q, bad %q", ok.fatal, bad.fatal)
	}
}
