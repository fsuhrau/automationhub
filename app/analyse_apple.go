package app

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// analyseAPP reads an unpacked .app folder
func (a *analyser) analyseAPP(info *BundleInfo) error {
	mac, err := readAppleFolder(a.appPath, info)
	if err != nil {
		return err
	}
	if mac {
		err = a.setMacFolder(info)
	} else {
		a.parameter.Platform = "ios"
		a.parameter.Name = info.DisplayName
	}
	if err != nil {
		return err
	}
	a.parameter.Version = info.Build
	a.parameter.Identifier = info.Identifier
	return nil
}

func (a *analyser) setMacFolder(info *BundleInfo) error {
	a.parameter.Platform = "macos"
	a.parameter.Name = filepath.Base(a.appPath)
	executable := info.Apple.Executable
	if executable == "" {
		file, err := getFirstFile(filepath.Join(a.appPath, "Contents", "MacOS"))
		if err != nil {
			return err
		}
		executable = "Contents/MacOS/" + filepath.Base(file)
	}
	a.parameter.App.Executable = &ExecutableParams{Executable: executable}
	return nil
}

func getFirstFile(dirPath string) (string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if !file.IsDir() {
			return filepath.Join(dirPath, file.Name()), nil
		}
	}
	return "", fmt.Errorf("no file found in %s", dirPath)
}

// analyseAppleArchive reads an ipa or a zipped .app without unpacking it
func (a *analyser) analyseAppleArchive(info *BundleInfo) error {
	r, err := zip.OpenReader(a.appPath)
	if err != nil {
		return err
	}
	defer closeQuietly(r)
	root, err := readAppleArchive(r.File, info)
	if err != nil {
		return err
	}
	appRoot := root
	if strings.HasSuffix(root, ".app/Contents/") {
		appRoot = a.setMacArchive(r.File, root, info)
	} else {
		a.parameter.Platform = "ios"
		a.parameter.Name = info.DisplayName
	}
	info.Engine = detectEngine(r.File, root)
	info.Contents = readContents(r.File, appRoot)
	// the build number tells builds apart, as before
	a.parameter.Version = info.Build
	if a.parameter.Version == "" {
		a.parameter.Version = info.Version
	}
	a.parameter.Identifier = info.Identifier
	return nil
}

// setMacArchive names the parameter after the Mac app and returns the folder
// of the app, which is above the root (Contents/)
func (a *analyser) setMacArchive(files []*zip.File, root string, info *BundleInfo) string {
	app := strings.TrimSuffix(root, "Contents/")
	a.parameter.Platform = "macos"
	a.parameter.Name = filepath.Base(strings.TrimSuffix(app, "/"))
	executable := info.Apple.Executable
	if executable == "" {
		executable = firstMacExecutable(files, root)
	}
	a.parameter.App.Executable = &ExecutableParams{Executable: executable}
	return app
}

// firstMacExecutable is the first file directly in Contents/MacOS
func firstMacExecutable(files []*zip.File, root string) string {
	for _, f := range files {
		rel, ok := strings.CutPrefix(f.Name, root+"MacOS/")
		if ok && rel != "" && !strings.Contains(rel, "/") {
			return "Contents/MacOS/" + rel
		}
	}
	return ""
}
