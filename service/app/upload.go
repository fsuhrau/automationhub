package app

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

// UploadFile is a build file that was sent in.
type UploadFile struct {
	Name    string
	Content io.Reader
}

// Upload saves the upload into the bundle folder and stores it as a bundle
// of the app.
func (s *Service) Upload(ctx context.Context, sc Scope, file UploadFile) (*models.AppBinary, error) {
	if _, err := s.repo.App(ctx, sc); err != nil {
		return nil, err
	}
	path, err := s.files.Save(file.Name, file.Content)
	if err != nil {
		return nil, fmt.Errorf("upload file err: %w", err)
	}
	return s.store(ctx, sc, path)
}

// StoreBundle analyses an uploaded app file, moves it into the bundle folder
// and creates the bundle entry. The file must be inside the folder. When the
// same build exists already, that bundle comes back with ErrBundleExists; an
// archived one is restored from the file instead.
func (s *Service) StoreBundle(ctx context.Context, sc Scope, filePath string) (*models.AppBinary, error) {
	if _, err := s.repo.App(ctx, sc); err != nil {
		return nil, err
	}
	return s.store(ctx, sc, filePath)
}

// store is StoreBundle for an app the scope was checked of.
func (s *Service) store(ctx context.Context, sc Scope, filePath string) (*models.AppBinary, error) {
	params, err := s.analyser.Analyse(filePath)
	if err != nil {
		_ = s.files.Remove(filePath)
		return nil, fmt.Errorf("unable to analyse app: %w", err)
	}
	existing, err := s.repo.BundleByHash(ctx, sc, params.App.Hash)
	if err == nil {
		return s.known(ctx, existing, filePath)
	}
	if apperr.KindOf(err) != apperr.KindNotFound {
		return nil, err
	}
	return s.create(ctx, filePath, newBundle(sc.AppID, filePath, params))
}

// known handles an upload of a file the app has a bundle of already.
func (s *Service) known(ctx context.Context, existing *models.AppBinary, filePath string) (*models.AppBinary, error) {
	if existing.Archived() {
		return s.restore(ctx, existing, filePath)
	}
	_ = s.files.Remove(filePath)
	return existing, ErrBundleExists
}

// restore puts the file of an archived bundle back, from an upload of the
// same file; the bundle keeps its id, tags and runs.
func (s *Service) restore(ctx context.Context, b *models.AppBinary, filePath string) (*models.AppBinary, error) {
	target := s.files.Path(b.AppPath)
	if target == "" {
		_ = s.files.Remove(filePath)
		return nil, fmt.Errorf("archived bundle %d has no file path to restore to", b.ID)
	}
	if err := s.files.Move(filePath, target); err != nil {
		return nil, fmt.Errorf("unable to store file: %w", err)
	}
	b.ArchivedAt = nil
	if err := s.repo.SaveArchivedAt(ctx, b); err != nil {
		return nil, err
	}
	_ = s.files.Remove(s.iconFile(b.AppPath))
	return b, nil
}

// create moves the file to the name the bundle has and stores the bundle.
func (s *Service) create(ctx context.Context, filePath string, b models.AppBinary) (*models.AppBinary, error) {
	if err := s.files.Move(filePath, s.files.Path(b.AppPath)); err != nil {
		return nil, fmt.Errorf("unable to store file: %w", err)
	}
	if err := s.repo.CreateBundle(ctx, &b); err != nil {
		return nil, fmt.Errorf("unable to create app bundle: %w", err)
	}
	return &b, nil
}

// newBundle is the bundle entry of an analysed file; its file is stored as
// name_hash.ext.
func newBundle(appID uint, filePath string, params *hubapp.Parameter) models.AppBinary {
	filename := filepath.Base(filePath)
	ext := filepath.Ext(filename)
	appPath := fmt.Sprintf("%s_%s%s", strings.TrimSuffix(filename, ext), params.App.Hash, ext)
	b := models.AppBinary{
		AppPath:    appPath,
		Hash:       params.App.Hash,
		Name:       params.Name,
		Identifier: params.Identifier,
		AppID:      appID,
		Version:    params.Version,
		Platform:   params.Platform,
		Additional: params.App.Additional,
		Size:       params.App.Size,
		Info:       params.Info,
	}
	if params.App.Android != nil {
		b.Android.LaunchActivity = params.App.Android.LaunchActivity
	}
	if params.App.Executable != nil {
		b.Executable.Executable = params.App.Executable.Executable
	}
	return b
}
