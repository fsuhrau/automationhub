package appfixture

import (
	"errors"
	"testing"
)

// stopRecorder is a testing.TB that records Fatal instead of ending the test.
type stopRecorder struct {
	testing.TB
	fatal bool
}

func (s *stopRecorder) Helper()      {}
func (s *stopRecorder) Fatal(...any) { s.fatal = true }

func TestMustStopsTheTestOnlyOnAnError(t *testing.T) {
	ok, bad := &stopRecorder{}, &stopRecorder{}
	must(ok, nil)
	must(bad, errors.New("boom"))
	if ok.fatal || !bad.fatal {
		t.Errorf("fatal after nil %v, after an error %v; want false and true", ok.fatal, bad.fatal)
	}
}
