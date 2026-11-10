package app

import (
	"archive/zip"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/tools/androidsdk"
	"github.com/fsuhrau/automationhub/tools/exec"
)

type analyser struct {
	appPath   string
	parameter Parameter
}

func NewAnalyser(appPath string) *analyser {
	return &analyser{
		appPath: appPath,
		parameter: Parameter{
			App: &AppParams{AppPath: appPath},
		},
	}
}

func (a *analyser) GetParameter() *Parameter {
	return &a.parameter
}

// AnalyseFile reads what the app is and hashes the file
func (a *analyser) AnalyseFile() error {
	if err := a.extractAppInfos(); err != nil {
		return err
	}
	stat, err := os.Stat(a.appPath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return nil
	}
	f, err := os.Open(a.appPath)
	if err != nil {
		return err
	}
	defer f.Close()
	hash := sha1.New()
	n, err := io.Copy(hash, f)
	if err != nil {
		return err
	}
	a.parameter.App.Size = int(n)
	a.parameter.App.Hash = fmt.Sprintf("%x", hash.Sum(nil))
	return nil
}

// ReadInfo reads what the app is without hashing the file, for a stored
// bundle that is read again
func (a *analyser) ReadInfo() error {
	return a.extractAppInfos()
}

// badging runs "aapt dump badging"; replaced in tests
var badging = func(apkPath string) ([]byte, error) {
	aapt, err := androidsdk.AAPT()
	if err != nil {
		return nil, err
	}
	return exec.NewCommand(aapt, "dump", "badging", apkPath).Output()
}

func (a *analyser) extractAppInfos() error {
	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(a.appPath)), ".")
	info := &BundleInfo{Format: format, ReadAt: time.Now().UTC()}
	a.parameter.Info = info
	var err error
	switch format {
	case "apk":
		err = a.analyseAPK(info)
	case "app":
		err = a.analyseAPP(info)
	case "ipa", "zip":
		err = a.analyseAppleArchive(info)
	}
	return err
}

// analyseAPP reads an unpacked .app folder
func (a *analyser) analyseAPP(info *BundleInfo) error {
	mac, err := readAppleFolder(a.appPath, info)
	if err != nil {
		return err
	}
	a.parameter.Platform = "ios"
	if mac {
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
	} else {
		a.parameter.Name = info.DisplayName
	}
	a.parameter.Version = info.Build
	a.parameter.Identifier = info.Identifier
	return nil
}

// analyseAppleArchive reads an ipa or a zipped .app without unpacking it
func (a *analyser) analyseAppleArchive(info *BundleInfo) error {
	r, err := zip.OpenReader(a.appPath)
	if err != nil {
		return err
	}
	defer r.Close()
	root, err := readAppleArchive(r.File, info)
	if err != nil {
		return err
	}
	mac := strings.HasSuffix(root, ".app/Contents/")
	a.parameter.Platform = "ios"
	if mac {
		a.parameter.Platform = "macos"
		app := strings.TrimSuffix(root, "Contents/")
		a.parameter.Name = filepath.Base(strings.TrimSuffix(app, "/"))
		executable := info.Apple.Executable
		if executable == "" {
			for _, f := range r.File {
				rel := strings.TrimPrefix(f.Name, root+"MacOS/")
				if rel != f.Name && rel != "" && !strings.Contains(rel, "/") {
					executable = "Contents/MacOS/" + rel
					break
				}
			}
		}
		a.parameter.App.Executable = &ExecutableParams{Executable: executable}
		info.Engine = detectEngine(r.File, root)
		info.Contents = readContents(r.File, app)
	} else {
		a.parameter.Name = info.DisplayName
		info.Engine = detectEngine(r.File, root)
		info.Contents = readContents(r.File, root)
	}
	// the build number tells builds apart, as before
	a.parameter.Version = info.Build
	if a.parameter.Version == "" {
		a.parameter.Version = info.Version
	}
	a.parameter.Identifier = info.Identifier
	return nil
}

func (a *analyser) analyseAPK(info *BundleInfo) error {
	a.parameter.Platform = "android"
	output, err := badging(a.appPath)
	if err != nil {
		return err
	}
	android, icons := parseBadging(output)
	info.Android = android
	info.Identifier = android.PackageName
	info.Version = android.VersionName
	info.Build = android.VersionCode
	info.DisplayName = android.Label
	a.parameter.Identifier = android.PackageName
	a.parameter.Version = android.VersionName
	a.parameter.Name = android.Label
	if android.LaunchActivity != "" {
		a.parameter.App.Android = &AndroidParams{LaunchActivity: android.LaunchActivity}
	}

	f, err := os.Open(a.appPath)
	if err != nil {
		return err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	r, err := zip.NewReader(f, stat.Size())
	if err != nil {
		info.Warnings = append(info.Warnings, fmt.Sprintf("unable to read the apk as zip: %v", err))
		return nil
	}
	info.Icon = apkIcon(r.File, icons)
	info.Engine = detectEngine(r.File, "")
	info.Contents = readContents(r.File, "")
	signing, err := readAPKSigning(f, stat.Size(), r.File)
	if err != nil {
		info.Warnings = append(info.Warnings, fmt.Sprintf("unable to read the signature: %v", err))
	}
	if signing != nil && (len(signing.Schemes) > 0 || len(signing.Certificates) > 0) {
		info.Signing = signing
	}
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

func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
