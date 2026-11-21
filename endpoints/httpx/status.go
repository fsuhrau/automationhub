// Package httpx is the one place where the api turns errors and requests
// into HTTP: the status of an apperr kind, the error body, and the readers
// for path ids, the limit parameter and the project and app of a request.
package httpx

import (
	"net/http"

	"github.com/fsuhrau/automationhub/apperr"
)

// Status is the HTTP status of an error: 200 for nil, the status of the
// kind otherwise.
func Status(err error) int {
	switch apperr.KindOf(err) {
	case apperr.KindNone:
		return http.StatusOK
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindInvalid:
		return http.StatusBadRequest
	case apperr.KindConflict, apperr.KindBusy:
		return http.StatusConflict
	case apperr.KindPreconditionRequired:
		return http.StatusPreconditionRequired
	case apperr.KindUnsupported:
		return http.StatusNotImplemented
	case apperr.KindTimeout:
		return http.StatusGatewayTimeout
	case apperr.KindForbidden:
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}
