package runstore

import (
	"context"
	"strings"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
)

// ProjectProtocol loads a protocol of the project with its device and keys.
func (s *Store) ProjectProtocol(ctx context.Context, projectID, protocolID uint) (*models.TestProtocol, error) {
	var p models.TestProtocol
	err := s.session(ctx).Preload("Device").Preload("Keys").
		Joins("JOIN test_runs ON test_runs.id = test_protocols.test_run_id").
		Joins("JOIN tests ON tests.id = test_runs.test_id").
		Joins("JOIN apps ON apps.id = tests.app_id").
		Where("apps.project_id = ?", projectID).
		First(&p, "test_protocols.id = ?", protocolID).Error
	if err != nil {
		return nil, lookup(err, "protocol", apperr.NotFound("not found: protocol %d", protocolID))
	}
	return &p, nil
}

// ProtocolLogs pages the log entries of the query and counts all that match.
func (s *Store) ProtocolLogs(ctx context.Context, q run.LogQuery) ([]models.ProtocolEntry, int64, error) {
	query := s.session(ctx).Model(&models.ProtocolEntry{}).Where("test_protocol_id = ?", q.ProtocolID)
	if q.Level != "" {
		query = query.Where("level = ?", strings.ToLower(q.Level))
	}
	if q.Contains != "" {
		query = query.Where(`message LIKE ? ESCAPE '\'`, "%"+escapeLike(q.Contains)+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, storeerr.Map(err, "protocol logs")
	}
	var entries []models.ProtocolEntry
	err := query.Order("id").Offset(q.Offset).Limit(q.Limit).Find(&entries).Error
	return entries, total, storeerr.Map(err, "protocol logs")
}

// escapeLike makes %, _ and \ in user input match literally in a LIKE with
// ESCAPE '\' (SQLite and Postgres alike).
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(s)
}
