// Package runstore is the gorm repository of test runs and their protocols.
// It implements service/run.Repo and service/protocol.Repo and maps gorm
// errors to apperr.
package runstore

import (
	"context"
	"errors"

	"github.com/fsuhrau/automationhub/service/protocol"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

var (
	_ run.Repo      = (*Store)(nil)
	_ protocol.Repo = (*Store)(nil)
)

// appTests selects the tests of one app; histories only look at runs of
// these, so they never mix in another project's data.
const appTests = "(select t.id from tests t where t.app_id = ? and t.deleted_at is null)"

// Store reads runs and protocols.
type Store struct {
	db *gorm.DB
}

// New builds the store on the database.
func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) session(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}

// lookup classifies the error of a single-row load about what: a missing
// row is the error missing, worded as the api has always answered it;
// anything else goes through storeerr.
func lookup(err error, what string, missing error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return missing
	}
	return storeerr.Map(err, what)
}

// require runs the counting query and answers missing when it matches no
// row.
func require(query *gorm.DB, what string, missing error) error {
	var n int64
	if err := query.Count(&n).Error; err != nil {
		return storeerr.Map(err, what)
	}
	if n == 0 {
		return missing
	}
	return nil
}
