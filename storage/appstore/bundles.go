package appstore

import (
	"context"
	"fmt"

	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
)

// Bundles lists the bundles of the app, newest first.
func (s *Store) Bundles(ctx context.Context, sc appsvc.Scope) ([]models.AppBinary, error) {
	var found []models.AppBinary
	err := inScope(s.tx(ctx), "app_id", sc).Order("id desc").Find(&found).Error
	return found, storeerr.Map(err, "bundles")
}

// Bundle loads one bundle of the app.
func (s *Store) Bundle(ctx context.Context, ref appsvc.BundleRef) (*models.AppBinary, error) {
	var found models.AppBinary
	err := inScope(s.tx(ctx), "app_id", ref.Scope).First(&found, "id = ?", ref.ID).Error
	return &found, lookup(err, fmt.Sprintf("bundle %d", ref.ID))
}

// BundleByHash loads the bundle of the app that was made from the file with
// the hash.
func (s *Store) BundleByHash(ctx context.Context, sc appsvc.Scope, hash string) (*models.AppBinary, error) {
	var found models.AppBinary
	err := inScope(s.tx(ctx), "app_id", sc).First(&found, "hash = ?", hash).Error
	return &found, lookup(err, "bundle "+hash)
}

// CreateBundle inserts the bundle.
func (s *Store) CreateBundle(ctx context.Context, b *models.AppBinary) error {
	return storeerr.Map(s.tx(ctx).Create(b).Error, "bundle")
}

// DeleteBundle removes the bundle.
func (s *Store) DeleteBundle(ctx context.Context, b *models.AppBinary) error {
	return storeerr.Map(s.tx(ctx).Delete(b).Error, fmt.Sprintf("bundle %d", b.ID))
}

// SaveTags writes the tags of the bundle.
func (s *Store) SaveTags(ctx context.Context, b *models.AppBinary) error {
	err := s.tx(ctx).Model(b).Update("tags", b.Tags).Error
	return storeerr.Map(err, fmt.Sprintf("bundle %d", b.ID))
}

// SaveArchivedAt writes when the bundle was archived; nil brings it back.
func (s *Store) SaveArchivedAt(ctx context.Context, b *models.AppBinary) error {
	err := s.tx(ctx).Model(b).Update("archived_at", b.ArchivedAt).Error
	return storeerr.Map(err, fmt.Sprintf("bundle %d", b.ID))
}

// SaveRead writes the columns of the bundle that reading its file sets.
func (s *Store) SaveRead(ctx context.Context, b *models.AppBinary, columns []string) error {
	err := s.tx(ctx).Model(b).Select(columns).Updates(b).Error
	return storeerr.Map(err, fmt.Sprintf("bundle %d", b.ID))
}
