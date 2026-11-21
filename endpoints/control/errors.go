package control

import (
	"errors"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/utils/sync"
	"github.com/gin-gonic/gin"
)

type sentinelKind struct {
	target error
	kind   func(format string, args ...any) error
}

// sentinelKinds says which apperr kind each sentinel error of the control
// service and the devices has. The first match wins. The constructor is
// used with "%w", so the message and errors.Is of the cause stay intact.
func sentinelKinds() []sentinelKind {
	return []sentinelKind{
		{hubcontrol.ErrLeaseRequired, apperr.PreconditionRequired},
		{hubcontrol.ErrLeaseInvalid, apperr.Conflict},
		{hubcontrol.ErrDeviceBusy, apperr.Conflict},
		{hubcontrol.ErrAppNotConnected, apperr.Conflict},
		{device.ErrUnsupported, apperr.Unsupported},
		{sync.TimeoutError, apperr.Timeout},
	}
}

// classify gives a service error the apperr kind of its sentinel. Errors
// without a known sentinel stay as they are and end up as internal errors.
func classify(err error) error {
	for _, s := range sentinelKinds() {
		if errors.Is(err, s.target) {
			return s.kind("%w", err)
		}
	}
	return err
}

// failErr answers a failed service call.
func failErr(c *gin.Context, err error) {
	httpx.Fail(c, classify(err))
}

// bind reads the JSON body into v and answers 400 when it is invalid.
func bind(c *gin.Context, v any) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		httpx.Fail(c, apperr.Invalid("%w", err))
		return false
	}
	return true
}
