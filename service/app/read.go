package app

import (
	"context"
	"errors"
	"net/http"
	"os"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Icon is the app icon out of the bundle file, or the copy kept when the
// bundle was archived.
func (s *Service) Icon(ctx context.Context, ref BundleRef) (*Icon, error) {
	b, err := s.repo.Bundle(ctx, ref)
	if err != nil {
		return nil, err
	}
	if b.Archived() {
		return s.archivedIcon(b)
	}
	file := s.files.Path(b.AppPath)
	if b.Info == nil || b.Info.Icon == "" || file == "" {
		return nil, noIcon()
	}
	data, contentType, err := s.analyser.Icon(file, b.Info.Icon)
	if errors.Is(err, hubapp.ErrNoIcon) {
		return nil, noIcon()
	}
	if err != nil {
		return nil, err
	}
	return &Icon{Data: data, ContentType: contentType}, nil
}

func (s *Service) archivedIcon(b *models.AppBinary) (*Icon, error) {
	iconFile := s.iconFile(b.AppPath)
	if iconFile == "" {
		return nil, noIcon()
	}
	data, err := s.files.Read(iconFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil, noIcon()
	}
	if err != nil {
		return nil, err
	}
	return &Icon{Data: data, ContentType: http.DetectContentType(data)}, nil
}

func noIcon() error { return apperr.NotFound("%s", hubapp.ErrNoIcon) }

// ReadAgain reads the stored file of a bundle again, for bundles uploaded
// before the master read all of it. The file, its hash and the tags stay;
// what the file says about the app is updated.
func (s *Service) ReadAgain(ctx context.Context, ref BundleRef) (*BundleDetail, error) {
	b, err := s.repo.Bundle(ctx, ref)
	if err != nil {
		return nil, err
	}
	if b.Archived() {
		return nil, apperr.Conflict("bundle %d is archived; its file is gone", b.ID)
	}
	file := s.files.Path(b.AppPath)
	if file == "" {
		return nil, apperr.Conflict("bundle %d has no file", b.ID)
	}
	params, err := s.analyser.Read(file)
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInvalid, err, "unable to read the bundle")
	}
	if err := s.repo.SaveRead(ctx, b, applyRead(b, params)); err != nil {
		return nil, err
	}
	return s.detail(ctx, b)
}

// applyRead sets what the file says on the bundle and returns the columns
// that changed. A value the file doesn't say leaves the bundle's own.
func applyRead(b *models.AppBinary, params *hubapp.Parameter) []string {
	b.Info = params.Info
	columns := []string{"info"}
	set := func(column string, field *string, value string) {
		if value != "" {
			*field = value
			columns = append(columns, column)
		}
	}
	set("name", &b.Name, params.Name)
	set("identifier", &b.Identifier, params.Identifier)
	set("version", &b.Version, params.Version)
	set("platform", &b.Platform, params.Platform)
	if params.App.Android != nil {
		set("android_launch_activity", &b.Android.LaunchActivity, params.App.Android.LaunchActivity)
	}
	if params.App.Executable != nil {
		set("executable_executable", &b.Executable.Executable, params.App.Executable.Executable)
	}
	return columns
}
