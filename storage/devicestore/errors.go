package devicestore

import (
	"errors"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

// mapErr classifies a storage error about what ("device 3"). A missing
// record is NotFound with the plain message "device 3 not found", which the
// API has always answered with; everything else goes through storeerr.
func mapErr(err error, what string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.NotFound("%s not found", what)
	}
	return storeerr.Map(err, what)
}
