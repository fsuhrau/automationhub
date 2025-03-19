package app

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/fsuhrau/automationhub/tools/exec"
	"howett.net/plist"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	AndroidAPKInfosRegex = regexp.MustCompile(`package: name='([a-zA-Z0-9._-]+)' versionCode='([a-zA-Z0-9._-]+)' versionName='([a-zA-Z0-9._/-]+)'.*`)
	//  compileSdkVersion='(.*)' compileSdkVersionCodename='(.*)' <- TODO check if we can it also on older adb versions
	LaunchActivityRegex = regexp.MustCompile(`launchable-activity:\s+name='([a-zA-Z0-9._-]+)'\s+label='(.*)'\sicon='.*'`)
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

type analyser struct {
	appPath   string
	parameter Parameter
}

func NewAnalyser(appPath string) *analyser {
	return &analyser{
		appPath: appPath,
		parameter: Parameter{
			AppPath: appPath,
		},
	}
}

func (a *analyser) GetParameter() *Parameter {
	return &a.parameter
}

func (a *analyser) AnalyseFile() error {
	if err := a.extractAppInfos(); err != nil {
		return err
	}

	data, err := ioutil.ReadFile(a.appPath)

	a.parameter.Size = len(data)

	if err != nil {
		return err
	}
	a.parameter.Hash = fmt.Sprintf("%x", sha1.Sum(data))
	return nil
}

func (a *analyser) extractAppInfos() error {
	extension := filepath.Ext(a.appPath)
	if extension == ".apk" {
		return a.analyseAPK()
	}

	if extension == ".app" {
		return a.analyseAPP(a.appPath)
	}

	if extension == ".ipa" {
		return a.analyseIPA()
	}

	if extension == ".zip" {
		return a.analyseZIP()
	}
	return nil
}

func (a *analyser) analyseAPP(path string) error {

	// apple app
	iosAppInfoPlist := filepath.Join(path, "Info.plist")
	if fileExists(iosAppInfoPlist) {
		a.parameter.Platform = "ios"
		plistData, err := ioutil.ReadFile(iosAppInfoPlist)
		if err != nil {
			return err
		}

		plistContent := map[string]interface{}{}
		_, err = plist.Unmarshal(plistData, &plistContent)
		if err != nil {
			return err
		}
		/*
			if val, ok := plistContent["DTPlatformName"]; ok {
				a.parameter.Platform = val.(string)
			}
		*/
		if val, ok := plistContent["CFBundleVersion"]; ok {
			a.parameter.Version = val.(string)
		}
		if val, ok := plistContent["CFBundleIdentifier"]; ok {
			a.parameter.Identifier = val.(string)
		}
	}

	macAppInfoPlist := filepath.Join(path, "Contents/Info.plist")
	if fileExists(macAppInfoPlist) {
		a.parameter.Platform = "macos"

		plistData, err := ioutil.ReadFile(macAppInfoPlist)
		if err != nil {
			return err
		}

		a.parameter.Name = filepath.Base(path)

		executableDir := filepath.Join(path, "Contents", "MacOS")

		file, err := getFirstFile(executableDir)
		if err != nil {
			return err
		}

		a.parameter.LaunchActivity = "Contents/MacOS/" + filepath.Base(file)

		plistContent := map[string]interface{}{}
		_, err = plist.Unmarshal(plistData, &plistContent)
		if err != nil {
			return err
		}
		/*
			if val, ok := plistContent["DTSDKName"]; ok {
				a.parameter.Platform = val.(string)
			}
		*/
		if val, ok := plistContent["CFBundleVersion"]; ok {
			a.parameter.Version = val.(string)
		}
		if val, ok := plistContent["CFBundleIdentifier"]; ok {
			a.parameter.Identifier = val.(string)
		}
	}

	return nil
}

func (a *analyser) analyseIPA() error {

	return a.analyseAPP(a.appPath)
}

func getFirstDirectory(dirPath string) (string, error) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if file.IsDir() {
			return filepath.Join(dirPath, file.Name()), nil
		}
	}

	return "", fmt.Errorf("no directory found in %s", dirPath)
}

func getFirstFile(dirPath string) (string, error) {
	files, err := ioutil.ReadDir(dirPath)
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

func (a *analyser) analyseZIP() error {

	tmpAppDir := filepath.Join(os.TempDir(), "automation_hub")
	os.MkdirAll(tmpAppDir, os.ModePerm)
	files, err := Unzip(a.appPath, tmpAppDir)
	if err != nil {
		return err
	}
	_ = files

	firstDir, _ := getFirstDirectory(tmpAppDir)

	return a.analyseAPP(firstDir)
}

func (a *analyser) analyseAPK() error {
	// android app
	a.parameter.Platform = "android"
	cmd := exec.NewCommand("aapt", "dump", "badging", a.appPath)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if matches := AndroidAPKInfosRegex.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			a.parameter.Identifier = matches[0][1]
			a.parameter.Version = matches[0][3]
			// a.parameter.Additional = fmt.Sprintf("versionCode: %s compileSdkVersion %s", matches[0][2], matches[0][4])
			continue
		}
		if matches := LaunchActivityRegex.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			a.parameter.LaunchActivity = matches[0][1]
			a.parameter.Name = matches[0][2]
			continue
		}
	}
	return nil
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
