// Package storeerr maps the errors of the storage layer (gorm) to apperr
// kinds, so handlers and services only ever see apperr errors. The mapping
// lives here and not in apperr, which stays free of the gorm dependency.
package storeerr

import (
	"errors"

	"github.com/fsuhrau/automationhub/apperr"
	"gorm.io/gorm"
)

var rules = []struct { //nolint:gochecknoglobals // read-only lookup table (concept: lookup tables allowed)
	target error
	kind   apperr.Kind
	suffix string
}{
	{gorm.ErrRecordNotFound, apperr.KindNotFound, "not found"},
	{gorm.ErrDuplicatedKey, apperr.KindConflict, "exists already"},
	{gorm.ErrForeignKeyViolated, apperr.KindConflict, "is still in use"},
	{gorm.ErrInvalidData, apperr.KindInvalid, "has invalid data"},
	{gorm.ErrInvalidValue, apperr.KindInvalid, "has an invalid value"},
}

// Map classifies a storage error about what ("test run 3"): a missing
// record becomes NotFound ("test run 3 not found"), a duplicate or a
// violated reference Conflict, invalid data Invalid, and everything else
// Internal ("test run 3: <cause>"). nil stays nil and an error that already
// has a kind is returned as it is.
func Map(err error, what string) error {
	if err == nil {
		return nil
	}
	var known *apperr.Error
	if errors.As(err, &known) {
		return err
	}
	for _, r := range rules {
		if errors.Is(err, r.target) {
			return apperr.Wrap(r.kind, err, what+" "+r.suffix)
		}
	}
	return apperr.Wrap(apperr.KindInternal, err, what)
}
