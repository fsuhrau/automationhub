package app

import (
	"archive/zip"
	"bytes"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"howett.net/plist"
)

// appleRoot finds the folder of the app inside an ipa or a zipped .app:
// Payload/X.app/ for an ipa, X.app/ or X.app/Contents/ for a zip. The app
// with the shortest path wins, so plug-ins inside it are skipped.
func appleRoot(files []*zip.File) (root string, mac bool, err error) {
	best := ""
	for _, f := range files {
		name := f.Name
		if strings.HasPrefix(name, "__MACOSX/") {
			continue
		}
		var candidate string
		var isMac bool
		switch {
		case strings.HasSuffix(name, ".app/Contents/Info.plist"):
			candidate, isMac = strings.TrimSuffix(name, "Info.plist"), true
		case strings.HasSuffix(name, ".app/Info.plist"):
			candidate = strings.TrimSuffix(name, "Info.plist")
		default:
			continue
		}
		if best == "" || len(candidate) < len(best) {
			best, mac = candidate, isMac
		}
	}
	if best == "" {
		return "", false, errors.New("no .app with an Info.plist in the archive")
	}
	return best, mac, nil
}

func plistString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		switch s := v.(type) {
		case string:
			return s
		case uint64, int64, float64, bool:
			return fmt.Sprint(s)
		}
	}
	return ""
}

func plistStrings(m map[string]interface{}, key string) []string {
	var out []string
	switch v := m[key].(type) {
	case []interface{}:
		for _, e := range v {
			switch s := e.(type) {
			case string:
				out = append(out, s)
			case uint64, int64:
				out = append(out, fmt.Sprint(s))
			}
		}
	case map[string]interface{}:
		// UIRequiredDeviceCapabilities can be a dict of capability: required
		for k, b := range v {
			if req, ok := b.(bool); !ok || req {
				out = append(out, k)
			}
		}
		sort.Strings(out)
	case string:
		out = append(out, v)
	}
	return out
}

var deviceFamilies = map[string]string{"1": "iPhone", "2": "iPad", "3": "Apple TV", "4": "Apple Watch", "6": "Mac (Catalyst)", "7": "Apple Vision"}

// readInfoPlist fills the info from an app's Info.plist
func readInfoPlist(data []byte, info *BundleInfo) (map[string]interface{}, error) {
	content := map[string]interface{}{}
	if _, err := plist.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("Info.plist: %w", err)
	}
	apple := &AppleInfo{
		Platform:             plistString(content, "DTPlatformName"),
		MinOS:                plistString(content, "MinimumOSVersion"),
		SDK:                  plistString(content, "DTSDKName"),
		Xcode:                xcodeVersion(plistString(content, "DTXcode")),
		XcodeBuild:           plistString(content, "DTXcodeBuild"),
		BuildMachineOS:       plistString(content, "BuildMachineOSBuild"),
		BundleName:           plistString(content, "CFBundleName"),
		Executable:           plistString(content, "CFBundleExecutable"),
		RequiredCapabilities: plistStrings(content, "UIRequiredDeviceCapabilities"),
		Orientations:         plistStrings(content, "UISupportedInterfaceOrientations"),
		BackgroundModes:      plistStrings(content, "UIBackgroundModes"),
	}
	if apple.MinOS == "" {
		apple.MinOS = plistString(content, "LSMinimumSystemVersion")
	}
	if b, ok := content["UIFileSharingEnabled"].(bool); ok {
		apple.FileSharing = b
	}
	for _, f := range plistStrings(content, "UIDeviceFamily") {
		if name, ok := deviceFamilies[f]; ok {
			apple.DeviceFamily = append(apple.DeviceFamily, name)
		} else {
			apple.DeviceFamily = append(apple.DeviceFamily, f)
		}
	}
	if types, ok := content["CFBundleURLTypes"].([]interface{}); ok {
		for _, t := range types {
			if m, ok := t.(map[string]interface{}); ok {
				apple.URLSchemes = append(apple.URLSchemes, plistStrings(m, "CFBundleURLSchemes")...)
			}
		}
	}
	keys := make([]string, 0, len(content))
	for k := range content {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if strings.HasSuffix(k, "UsageDescription") {
			apple.UsageDescriptions = append(apple.UsageDescriptions, UsageDescription{Key: k, Text: plistString(content, k)})
		}
	}
	info.Apple = apple
	info.Identifier = plistString(content, "CFBundleIdentifier")
	info.Version = plistString(content, "CFBundleShortVersionString")
	info.Build = plistString(content, "CFBundleVersion")
	info.DisplayName = plistString(content, "CFBundleDisplayName")
	if info.DisplayName == "" {
		info.DisplayName = apple.BundleName
	}
	return content, nil
}

// xcodeVersion turns DTXcode "1520" into 15.2
func xcodeVersion(v string) string {
	if len(v) < 3 {
		return v
	}
	major := strings.TrimLeft(v[:len(v)-2], "0")
	minor := v[len(v)-2 : len(v)-1]
	patch := v[len(v)-1:]
	if patch == "0" {
		return major + "." + minor
	}
	return major + "." + minor + "." + patch
}

// iconNames lists the icon file names an Info.plist names
func iconNames(content map[string]interface{}) []string {
	var names []string
	for _, key := range []string{"CFBundleIcons", "CFBundleIcons~ipad"} {
		icons, ok := content[key].(map[string]interface{})
		if !ok {
			continue
		}
		if primary, ok := icons["CFBundlePrimaryIcon"].(map[string]interface{}); ok {
			names = append(names, plistStrings(primary, "CFBundleIconFiles")...)
			if n := plistString(primary, "CFBundleIconName"); n != "" {
				names = append(names, n)
			}
		}
	}
	names = append(names, plistStrings(content, "CFBundleIconFiles")...)
	if n := plistString(content, "CFBundleIconFile"); n != "" {
		names = append(names, n)
	}
	return names
}

// appleIcon picks the largest icon file of the app in the archive
func appleIcon(files []*zip.File, root string, mac bool, content map[string]interface{}) string {
	names := iconNames(content)
	if len(names) == 0 {
		return ""
	}
	dir := root
	if mac {
		dir = root + "Resources/"
	}
	best := ""
	var bestSize uint64
	for _, f := range files {
		if !strings.HasPrefix(f.Name, dir) || strings.Contains(strings.TrimPrefix(f.Name, dir), "/") {
			continue
		}
		base := strings.TrimPrefix(f.Name, dir)
		ext := strings.ToLower(filepath.Ext(base))
		if ext != ".png" && ext != ".icns" {
			continue
		}
		stem := strings.TrimSuffix(base, filepath.Ext(base))
		for _, n := range names {
			n = strings.TrimSuffix(n, filepath.Ext(n))
			if stem == n || strings.HasPrefix(stem, n+"@") || strings.HasPrefix(stem, n+"~") {
				if f.UncompressedSize64 > bestSize {
					best, bestSize = f.Name, f.UncompressedSize64
				}
			}
		}
	}
	return best
}

type provisioningPlist struct {
	Name                  string                 `plist:"Name"`
	UUID                  string                 `plist:"UUID"`
	AppIDName             string                 `plist:"AppIDName"`
	TeamName              string                 `plist:"TeamName"`
	TeamIdentifier        []string               `plist:"TeamIdentifier"`
	CreationDate          time.Time              `plist:"CreationDate"`
	ExpirationDate        time.Time              `plist:"ExpirationDate"`
	ProvisionedDevices    []string               `plist:"ProvisionedDevices"`
	ProvisionsAllDevices  bool                   `plist:"ProvisionsAllDevices"`
	Entitlements          map[string]interface{} `plist:"Entitlements"`
	DeveloperCertificates [][]byte               `plist:"DeveloperCertificates"`
}

// readProvisioning reads an embedded.mobileprovision (or
// embedded.provisionprofile on the Mac): a signed plist
func readProvisioning(data []byte) (*ProvisioningProfile, []Certificate, error) {
	start := bytes.Index(data, []byte("<?xml"))
	end := bytes.Index(data, []byte("</plist>"))
	if start < 0 || end < start {
		return nil, nil, errors.New("no plist in the provisioning profile")
	}
	var p provisioningPlist
	if _, err := plist.Unmarshal(data[start:end+len("</plist>")], &p); err != nil {
		return nil, nil, fmt.Errorf("provisioning profile: %w", err)
	}
	profile := &ProvisioningProfile{
		Name:         p.Name,
		UUID:         p.UUID,
		AppIDName:    p.AppIDName,
		TeamName:     p.TeamName,
		CreatedAt:    p.CreationDate,
		ExpiresAt:    p.ExpirationDate,
		AllDevices:   p.ProvisionsAllDevices,
		Devices:      p.ProvisionedDevices,
		Entitlements: p.Entitlements,
	}
	if len(p.TeamIdentifier) > 0 {
		profile.TeamID = p.TeamIdentifier[0]
	}
	getTaskAllow, _ := p.Entitlements["get-task-allow"].(bool)
	switch {
	case p.ProvisionsAllDevices:
		profile.Kind = "enterprise"
	case len(p.ProvisionedDevices) > 0 && getTaskAllow:
		profile.Kind = "development"
	case len(p.ProvisionedDevices) > 0:
		profile.Kind = "ad-hoc"
	default:
		profile.Kind = "app-store"
	}
	var certs []Certificate
	for _, der := range p.DeveloperCertificates {
		if c, err := x509.ParseCertificate(der); err == nil {
			certs = append(certs, certificateOf(c))
		}
	}
	return profile, certs, nil
}

// readAppleArchive reads an ipa or a zipped .app
func readAppleArchive(files []*zip.File, info *BundleInfo) (root string, err error) {
	root, mac, err := appleRoot(files)
	if err != nil {
		return "", err
	}
	plistPath := root + "Info.plist"
	data, err := readZipFile(findFile(files, plistPath), 16<<20)
	if err != nil {
		return "", err
	}
	content, err := readInfoPlist(data, info)
	if err != nil {
		return "", err
	}
	if mac && info.Apple.Platform == "" {
		info.Apple.Platform = "macosx"
	}
	if mac && info.Apple.Executable != "" {
		info.Apple.Executable = "Contents/MacOS/" + info.Apple.Executable
	}
	info.Icon = appleIcon(files, root, mac, content)

	profileName := "embedded.mobileprovision"
	if mac {
		profileName = "embedded.provisionprofile"
	}
	if f := findFile(files, root+profileName); f != nil {
		data, err := readZipFile(f, 4<<20)
		if err == nil {
			var certs []Certificate
			info.Apple.Profile, certs, err = readProvisioning(data)
			if len(certs) > 0 {
				info.Signing = &Signing{Schemes: []string{"profile"}, Certificates: certs}
			}
		}
		if err != nil {
			info.Warnings = append(info.Warnings, err.Error())
		}
	}

	seen := map[string]bool{}
	for _, f := range files {
		rel := strings.TrimPrefix(f.Name, root)
		if rel == f.Name {
			continue
		}
		for _, dir := range []string{"PlugIns/", "Frameworks/", "Extensions/"} {
			if !strings.HasPrefix(rel, dir) {
				continue
			}
			name := strings.SplitN(strings.TrimPrefix(rel, dir), "/", 2)[0]
			if name == "" || seen[dir+name] {
				continue
			}
			seen[dir+name] = true
			if dir == "Frameworks/" {
				info.Apple.Frameworks = append(info.Apple.Frameworks, name)
			} else if strings.HasSuffix(name, ".appex") || strings.HasSuffix(name, ".xctest") {
				info.Apple.Extensions = append(info.Apple.Extensions, name)
			}
		}
	}
	return root, nil
}

// readAppleFolder reads an unpacked .app on disk, as the analyser got it for
// local apps before archives were read directly
func readAppleFolder(path string, info *BundleInfo) (mac bool, err error) {
	plistPath := filepath.Join(path, "Info.plist")
	if _, err := os.Stat(plistPath); err != nil {
		plistPath, mac = filepath.Join(path, "Contents", "Info.plist"), true
	}
	data, err := os.ReadFile(plistPath)
	if err != nil {
		return mac, err
	}
	_, err = readInfoPlist(data, info)
	if err == nil && mac && info.Apple.Executable != "" {
		info.Apple.Executable = "Contents/MacOS/" + info.Apple.Executable
	}
	return mac, err
}
