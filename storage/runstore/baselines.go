package runstore

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/protocol"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

// candidates selects the protocols a protocol can be compared with: the
// same test name in another run of the app, with performance measured.
func (s *Store) candidates(ctx context.Context, appID uint, current *models.TestProtocol) *gorm.DB {
	return s.session(ctx).Model(&models.TestProtocol{}).
		Select("test_protocols.*").
		Joins("JOIN test_runs ON test_runs.id = test_protocols.test_run_id AND test_runs.deleted_at IS NULL").
		Joins("JOIN tests ON tests.id = test_runs.test_id").
		Joins("LEFT JOIN app_binaries ON app_binaries.id = test_runs.app_binary_id").
		Where("tests.app_id = ? AND test_protocols.test_name = ? AND test_protocols.test_run_id <> ?",
			appID, current.TestName, current.TestRunID).
		Where("exists (select 1 from protocol_performance_entries pe where pe.test_protocol_id = test_protocols.id and pe.deleted_at is null)")
}

func withBaselineRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("Device").Preload("TestRun").Preload("TestRun.AppBinary")
}

// Candidates lists the protocols the current one can be compared with,
// newest first.
func (s *Store) Candidates(ctx context.Context, q protocol.CandidateQuery) ([]models.TestProtocol, error) {
	query := s.candidates(ctx, q.AppID, q.Current)
	if !q.Filter.AnyDevice {
		query = onDevice(query, q.Current.DeviceID)
	}
	if q.Filter.Tag != "" {
		query = query.Where("(',' || app_binaries.tags || ',') LIKE ? ESCAPE '\\'", "%,"+escapeLike(q.Filter.Tag)+",%")
	}
	if q.Filter.Query != "" {
		query = matching(query, q.Filter.Query)
	}
	var found []models.TestProtocol
	err := withBaselineRelations(query).Order("test_protocols.id desc").Limit(q.Filter.Limit).Find(&found).Error
	return found, storeerr.Map(err, "baselines")
}

func onDevice(query *gorm.DB, deviceID *uint) *gorm.DB {
	if deviceID == nil {
		return query.Where("test_protocols.device_id IS NULL")
	}
	return query.Where("test_protocols.device_id = ?", *deviceID)
}

// matching keeps the candidates whose bundle name, version or tags, or
// whose run parameters contain the text, or whose run is the number.
func matching(query *gorm.DB, text string) *gorm.DB {
	like := "%" + escapeLike(text) + "%"
	cond := `LOWER(app_binaries.name) LIKE ? ESCAPE '\' OR LOWER(app_binaries.version) LIKE ? ESCAPE '\'` +
		` OR LOWER(app_binaries.tags) LIKE ? ESCAPE '\' OR LOWER(test_runs.parameter) LIKE ? ESCAPE '\'`
	args := []any{like, like, like, like}
	if id, err := strconv.Atoi(strings.TrimPrefix(text, "#")); err == nil {
		cond += " OR test_runs.id = ?"
		args = append(args, id)
	}
	return query.Where("("+cond+")", args...)
}

// Baseline loads the candidate in the run with its performance entries: on
// the device of the query, or else preferably on the same device as the
// current protocol.
func (s *Store) Baseline(ctx context.Context, q protocol.BaselineQuery) (*models.TestProtocol, error) {
	query := s.candidates(ctx, q.AppID, q.Current).Where("test_protocols.test_run_id = ?", q.RunID)
	if q.DeviceID != nil {
		query = query.Where("test_protocols.device_id = ?", *q.DeviceID)
	} else if q.Current.DeviceID != nil {
		// a number, so it can be part of the clause
		query = query.Order(fmt.Sprintf("CASE WHEN test_protocols.device_id = %d THEN 0 ELSE 1 END", *q.Current.DeviceID))
	}
	var found models.TestProtocol
	err := withBaselineRelations(query).Order("test_protocols.id desc").
		Preload("Performance", func(db *gorm.DB) *gorm.DB { return db.Order("id") }).
		First(&found).Error
	if err != nil {
		missing := apperr.NotFound("run %d has no performance of %s to compare with", q.RunID, q.Current.TestName)
		return nil, lookup(err, "baseline", missing)
	}
	return &found, nil
}

// BundleTags lists the distinct tag lists of the app's bundles.
func (s *Store) BundleTags(ctx context.Context, appID uint) ([]string, error) {
	var lists []string
	err := s.session(ctx).Model(&models.AppBinary{}).Where("app_id = ? AND tags <> ''", appID).
		Distinct().Pluck("tags", &lists).Error
	return lists, storeerr.Map(err, "bundle tags")
}
