package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
)

// tooLargeError is a body that is longer than the limit.
type tooLargeError struct{ limit int64 }

func (e tooLargeError) Error() string { return fmt.Sprintf("upload is larger than %d bytes", e.limit) }

// receiver writes bodies into fresh folders of dir; a body longer than limit
// is refused.
type receiver struct {
	dir   string
	limit int64
}

// tempFolder creates a fresh folder in dir, and dir if it is not there.
func tempFolder(dir string) (string, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}
	return os.MkdirTemp(dir, "upload-")
}

// receive writes the body as the file into a fresh folder and returns the path
// of the file and a function that removes the folder.
func (r receiver) receive(fileName string, body io.Reader) (string, func(), error) {
	folder, err := tempFolder(r.dir)
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(folder) }
	path := filepath.Join(folder, fileName)
	if err := r.write(path, body); err != nil {
		cleanup()
		return "", nil, err
	}
	return path, cleanup, nil
}

// write copies the body into a new file; an empty or oversized body is an
// error.
func (r receiver) write(path string, body io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, io.LimitReader(body, r.limit+1))
	err = errors.Join(err, f.Close())
	switch {
	case err != nil:
		return fmt.Errorf("receive file: %w", err)
	case n > r.limit:
		return tooLargeError{r.limit}
	case n == 0:
		return errors.New("the upload is empty")
	}
	return nil
}

// uploadFile is a file that is sent in as a bundle of an app.
type uploadFile struct {
	scope appsvc.Scope
	name  string
	body  io.Reader
}

// storeUpload receives the body as a file and adds it as a bundle of the app.
// exists is true when the same build was stored before; bundle is that one.
func (d Deps) storeUpload(ctx context.Context, f uploadFile) (bundle *models.AppBinary, exists bool, err error) {
	path, cleanup, err := receiver{dir: d.StoreDir, limit: d.MaxUpload}.receive(f.name, f.body)
	if err != nil {
		return nil, false, err
	}
	defer cleanup()
	bundle, err = d.Apps.StoreBundle(ctx, f.scope, path)
	if errors.Is(err, appsvc.ErrBundleExists) {
		return bundle, true, nil
	}
	return bundle, false, err
}
