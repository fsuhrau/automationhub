package control

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/device"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/utils/sync"
	"github.com/gin-gonic/gin"
)

// legacyStatusFor is the statusFor the handler used before the apperr kinds.
func legacyStatusFor(err error) int {
	switch {
	case errors.Is(err, hubcontrol.ErrLeaseRequired):
		return http.StatusPreconditionRequired
	case errors.Is(err, hubcontrol.ErrLeaseInvalid), errors.Is(err, hubcontrol.ErrDeviceBusy), errors.Is(err, hubcontrol.ErrAppNotConnected):
		return http.StatusConflict
	case errors.Is(err, device.ErrUnsupported):
		return http.StatusNotImplemented
	case errors.Is(err, sync.ErrTimeout):
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}

func TestFailErrKeepsTheStatusOfEverySentinel(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"lease required", hubcontrol.ErrLeaseRequired, http.StatusPreconditionRequired},
		{"lease invalid", hubcontrol.ErrLeaseInvalid, http.StatusConflict},
		{"device busy", hubcontrol.ErrDeviceBusy, http.StatusConflict},
		{"app not connected", hubcontrol.ErrAppNotConnected, http.StatusConflict},
		{"unsupported", device.ErrUnsupported, http.StatusNotImplemented},
		{"timeout", sync.ErrTimeout, http.StatusGatewayTimeout},
		{"wrapped sentinel", fmt.Errorf("tap: %w", hubcontrol.ErrLeaseRequired), http.StatusPreconditionRequired},
		{"other error", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			failErr(c, tt.err)
			if rec.Code != tt.want || rec.Code != legacyStatusFor(tt.err) {
				t.Fatalf("status %d, want %d (legacy %d)", rec.Code, tt.want, legacyStatusFor(tt.err))
			}
			if want := `{"message":` + fmt.Sprintf("%q", tt.err.Error()) + `}`; rec.Body.String() != want {
				t.Fatalf("body %s, want %s", rec.Body, want)
			}
		})
	}
}

func TestClassifyKeepsTheCause(t *testing.T) {
	if err := classify(hubcontrol.ErrDeviceBusy); !errors.Is(err, hubcontrol.ErrDeviceBusy) {
		t.Fatalf("cause lost: %v", err)
	}
	other := errors.New("boom")
	if classify(other) != other {
		t.Fatal("an unknown error must stay as it is")
	}
}

func TestBindRejectsInvalidJSON(t *testing.T) {
	e := setup(t)
	rec := e.call(t, "", http.MethodPost, "1/key", map[string]string{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d %s", rec.Code, rec.Body)
	}
}
