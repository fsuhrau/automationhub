package app

import (
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
	// badging runs "aapt dump badging"; replaced in tests
	badging func(apkPath string) ([]byte, error)
}

func NewAnalyser(appPath string) *analyser {
	return &analyser{
		appPath: appPath,
		parameter: Parameter{
			App: &AppParams{AppPath: appPath},
		},
		badging: runAAPTBadging,
	}
}

func runAAPTBadging(apkPath string) ([]byte, error) {
	aapt, err := androidsdk.AAPT()
	if err != nil {
		return nil, err
	}
	return exec.NewCommand(aapt, "dump", "badging", apkPath).Output()
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
	size, hash, err := hashFile(a.appPath)
	if err != nil {
		return err
	}
	a.parameter.App.Size = size
	a.parameter.App.Hash = hash
	return nil
}

// hashFile returns the size and the SHA-1 of a file
func hashFile(path string) (int, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, "", err
	}
	defer f.Close()
	hash := sha1.New()
	n, err := io.Copy(hash, f)
	if err != nil {
		return 0, "", err
	}
	return int(n), fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// ReadInfo reads what the app is without hashing the file, for a stored
// bundle that is read again
func (a *analyser) ReadInfo() error {
	return a.extractAppInfos()
}

func (a *analyser) extractAppInfos() error {
	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(a.appPath)), ".")
	info := &BundleInfo{Format: format, ReadAt: time.Now().UTC()}
	a.parameter.Info = info
	switch format {
	case "apk":
		return a.analyseAPK(info)
	case "app":
		return a.analyseAPP(info)
	case "ipa", "zip":
		return a.analyseAppleArchive(info)
	}
	return nil
}
