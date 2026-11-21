package runstore

import (
	"context"

	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
)

// ProtocolDigest summarises the protocol; it lists at most limit errors.
func (s *Store) ProtocolDigest(ctx context.Context, protocolID uint, limit int) (run.Digest, error) {
	var d run.Digest
	var err error
	if d.Entries, err = s.entryCount(ctx, protocolID); err != nil {
		return run.Digest{}, err
	}
	if d.Errors, err = s.errorEntries(ctx, protocolID, limit); err != nil {
		return run.Digest{}, err
	}
	if d.Performance, err = s.performance(ctx, protocolID); err != nil {
		return run.Digest{}, err
	}
	if d.Children, err = s.children(ctx, protocolID); err != nil {
		return run.Digest{}, err
	}
	d.Data, err = s.DataEntries(ctx, protocolID)
	return d, err
}

func (s *Store) entryCount(ctx context.Context, protocolID uint) (int64, error) {
	var n int64
	err := s.session(ctx).Model(&models.ProtocolEntry{}).Where("test_protocol_id = ?", protocolID).Count(&n).Error
	return n, storeerr.Map(err, "protocol entries")
}

func (s *Store) errorEntries(ctx context.Context, protocolID uint, limit int) ([]models.ProtocolEntry, error) {
	var found []models.ProtocolEntry
	err := s.session(ctx).Where("test_protocol_id = ? AND level IN ?", protocolID, []string{"error", "exception"}).
		Order("id").Limit(limit).Find(&found).Error
	return found, storeerr.Map(err, "protocol errors")
}

func (s *Store) performance(ctx context.Context, protocolID uint) (run.Performance, error) {
	var perf run.Performance
	err := s.session(ctx).Model(&models.ProtocolPerformanceEntry{}).Where("test_protocol_id = ?", protocolID).
		Select("COUNT(*) AS samples, COALESCE(AVG(fps),0) AS avg_fps, COALESCE(MIN(fps),0) AS min_fps, " +
			"COALESCE(AVG(mem),0) AS avg_mem, COALESCE(MAX(mem),0) AS max_mem, COALESCE(AVG(cpu),0) AS avg_cpu").
		Scan(&perf).Error
	return perf, storeerr.Map(err, "protocol performance")
}

func (s *Store) children(ctx context.Context, protocolID uint) ([]models.TestProtocol, error) {
	var found []models.TestProtocol
	err := s.session(ctx).Preload("Device").Preload("Keys").
		Where("parent_test_protocol_id = ?", protocolID).Order("id").Find(&found).Error
	return found, storeerr.Map(err, "child protocols")
}

// DataEntries lists the entries of the protocol that carry data, by id.
func (s *Store) DataEntries(ctx context.Context, protocolID uint) ([]models.ProtocolEntry, error) {
	var found []models.ProtocolEntry
	err := s.session(ctx).Where("test_protocol_id = ? AND data <> ''", protocolID).Order("id").Find(&found).Error
	return found, storeerr.Map(err, "protocol data")
}
