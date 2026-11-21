package app

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip extracts the archive into dest and returns the paths it wrote
func Unzip(src string, dest string) ([]string, error) {
	var filenames []string
	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer closeQuietly(r)
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}
		filenames = append(filenames, fpath)
		if err := extractEntry(f, fpath); err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// extractEntry writes a folder or a file of the archive to fpath
func extractEntry(f *zip.File, fpath string) error {
	if f.FileInfo().IsDir() {
		return os.MkdirAll(fpath, os.ModePerm)
	}
	if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
		return err
	}
	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer closeQuietly(outFile)
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer closeQuietly(rc)
	_, err = io.Copy(outFile, rc)
	return err
}
