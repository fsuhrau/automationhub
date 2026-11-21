// Package teststore is the gorm repository of the test definitions. It
// implements service/test.Repo and maps gorm errors to apperr.
package teststore

import "gorm.io/gorm"

// LeafProtocols limits a query on test_protocols to the protocols that wrap
// no other protocol: the tests themselves. A runner protocol that wraps the
// tests of one run is left out, so it is not counted twice. The query must
// select from test_protocols without an alias.
func LeafProtocols(db *gorm.DB) *gorm.DB {
	return db.Where("NOT EXISTS (SELECT 1 FROM test_protocols c WHERE c.parent_test_protocol_id = test_protocols.id AND c.deleted_at IS NULL)")
}

// InRuns returns the scope that limits a query on test_protocols to the
// protocols of the runs with the ids.
func InRuns(runIDs []uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("test_protocols.test_run_id IN ?", runIDs)
	}
}
