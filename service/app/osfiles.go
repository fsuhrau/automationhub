package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	hubapp "github.com/fsuhrau/automationhub/app"
)

// OSFiles keeps the bundle files in a folder of the file system.
type OSFiles struct {
	root string
}

// NewOSFiles keeps the files below the root folder.
func NewOSFiles(root string) *OSFiles {
	return &OSFiles{root: filepath.Clean(root)}
}

// Path is the file below the root, or "" when name is empty or would leave
// the folder. An empty name used to make a delete remove the whole folder,
// with every bundle in it.
func (f *OSFiles) Path(name string) string {
	if name == "" {
		return ""
	}
	file := filepath.Join(f.root, name)
	if !strings.HasPrefix(file, f.root+string(filepath.Separator)) {
		return ""
	}
	return file
}

// Save writes the upload below the root under its base name.
func (f *OSFiles) Save(name string, src io.Reader) (string, error) {
	if err := os.MkdirAll(f.root, os.ModePerm); err != nil {
		return "", err
	}
	path := filepath.Join(f.root, filepath.Base(name))
	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(dst, src)
	if cerr := dst.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	return path, nil
}

// Move renames the file.
func (f *OSFiles) Move(from, to string) error { return os.Rename(from, to) }

// Remove deletes the file or folder.
func (f *OSFiles) Remove(path string) error { return os.RemoveAll(path) }

// Read returns the content of the file.
func (f *OSFiles) Read(path string) ([]byte, error) { return os.ReadFile(path) }

// Write stores the content in the file.
func (f *OSFiles) Write(path string, data []byte) error { return os.WriteFile(path, data, 0o644) }

// FileAnalyser reads bundle files with the app analysers.
type FileAnalyser struct{}

// Analyse reads and hashes the file.
func (FileAnalyser) Analyse(path string) (*hubapp.Parameter, error) {
	a := hubapp.NewAnalyser(path)
	if err := a.AnalyseFile(); err != nil {
		return nil, err
	}
	return a.GetParameter(), nil
}

// Read reads the file without hashing it.
func (FileAnalyser) Read(path string) (*hubapp.Parameter, error) {
	a := hubapp.NewAnalyser(path)
	if err := a.ReadInfo(); err != nil {
		return nil, err
	}
	return a.GetParameter(), nil
}

// Icon extracts the icon at the path inside the bundle.
func (FileAnalyser) Icon(path, iconPath string) ([]byte, string, error) {
	return hubapp.Icon(path, iconPath)
}
