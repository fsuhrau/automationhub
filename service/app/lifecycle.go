package app

import (
	"context"
	"fmt"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

// iconFile is where the icon of an archived bundle is kept, next to where
// its file was; "" when the bundle has no file path.
func (s *Service) iconFile(appPath string) string {
	file := s.files.Path(appPath)
	if file == "" {
		return ""
	}
	return file + ".icon"
}

// DeleteBundle removes the bundle and its file, and nothing else: a bundle
// without a file path leaves the storage folder alone.
func (s *Service) DeleteBundle(ctx context.Context, ref BundleRef) error {
	b, err := s.repo.Bundle(ctx, ref)
	if err != nil {
		return err
	}
	if file := s.files.Path(b.AppPath); file != "" {
		if err := s.files.Remove(file); err != nil {
			return err
		}
		_ = s.files.Remove(s.iconFile(b.AppPath))
	}
	return s.repo.DeleteBundle(ctx, b)
}

// Archive removes the file of a bundle and keeps everything read from it, so
// the runs with it stay comparable. An archived bundle can't be installed or
// run anymore; uploading the same file again brings it back. Archiving twice
// removes what is left of the file.
func (s *Service) Archive(ctx context.Context, ref BundleRef) (*BundleDetail, error) {
	b, err := s.repo.Bundle(ctx, ref)
	if err != nil {
		return nil, err
	}
	if !b.Archived() {
		if err := s.markArchived(ctx, b); err != nil {
			return nil, err
		}
	}
	if file := s.files.Path(b.AppPath); file != "" {
		if err := s.files.Remove(file); err != nil {
			return nil, apperr.Wrap(apperr.KindInternal, err, fmt.Sprintf("bundle %d is archived, but its file stays", b.ID))
		}
	}
	return s.detail(ctx, b)
}

// markArchived keeps a copy of the icon and sets the archive time. The time
// is set before the file goes, so no run or install starts with a file being
// removed.
func (s *Service) markArchived(ctx context.Context, b *models.AppBinary) error {
	inUse, err := s.inUse(ctx, b.ID)
	if err != nil {
		return err
	}
	if inUse {
		return apperr.Conflict("a test run uses bundle %d; archive it when the run is over", b.ID)
	}
	s.keepIcon(b)
	now := time.Now()
	b.ArchivedAt = &now
	return s.repo.SaveArchivedAt(ctx, b)
}

// keepIcon copies the icon out of the file, for the bundle pages.
func (s *Service) keepIcon(b *models.AppBinary) {
	file := s.files.Path(b.AppPath)
	if file == "" || b.Info == nil || b.Info.Icon == "" {
		return
	}
	if data, _, err := s.analyser.Icon(file, b.Info.Icon); err == nil {
		_ = s.files.Write(s.iconFile(b.AppPath), data)
	}
}
