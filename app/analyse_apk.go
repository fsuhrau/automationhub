package app

import (
	"archive/zip"
	"fmt"
	"os"
)

func (a *analyser) analyseAPK(info *BundleInfo) error {
	a.parameter.Platform = "android"
	output, err := a.badging(a.appPath)
	if err != nil {
		return err
	}
	icons := a.applyBadging(output, info)
	return a.readAPKArchive(info, icons)
}

// applyBadging takes what "aapt dump badging" says into the info and the
// parameter, and returns the icons it names by density
func (a *analyser) applyBadging(output []byte, info *BundleInfo) map[int]string {
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
	return icons
}

// readAPKArchive reads the icon, the engine, the contents and the signing of
// the apk; an apk that is no zip is a warning, not a failure
func (a *analyser) readAPKArchive(info *BundleInfo, icons map[int]string) error {
	stat, err := os.Stat(a.appPath)
	if err != nil {
		return err
	}
	f, err := os.Open(a.appPath)
	if err != nil {
		return err
	}
	defer closeQuietly(f)
	r, err := zip.NewReader(f, stat.Size())
	if err != nil {
		info.Warnings = append(info.Warnings, fmt.Sprintf("unable to read the apk as zip: %v", err))
		return nil
	}
	info.Icon = apkIcon(r.File, icons)
	info.Engine = detectEngine(r.File, "")
	info.Contents = readContents(r.File, "")
	signing, err := readAPKSigning(f, stat.Size(), r.File)
	applySigning(info, signing, err)
	return nil
}

// applySigning adds the signing to the info when there is any; an error is a
// warning
func applySigning(info *BundleInfo, signing *Signing, err error) {
	if err != nil {
		info.Warnings = append(info.Warnings, fmt.Sprintf("unable to read the signature: %v", err))
	}
	if signing != nil && (len(signing.Schemes) > 0 || len(signing.Certificates) > 0) {
		info.Signing = signing
	}
}
