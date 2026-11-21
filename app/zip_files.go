package app

import (
	"archive/zip"
	"io"
)

// closeQuietly closes what was only read; there is nothing to do when that fails
func closeQuietly(c io.Closer) {
	_ = c.Close()
}

func findFile(files []*zip.File, name string) *zip.File {
	for _, f := range files {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func readZipFile(f *zip.File, limit int64) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer closeQuietly(rc)
	return io.ReadAll(io.LimitReader(rc, limit))
}
