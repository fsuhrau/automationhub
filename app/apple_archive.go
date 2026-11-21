package app

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
)

// appleBundle is the app inside an ipa or a zipped .app
type appleBundle struct {
	files []*zip.File
	root  string
	mac   bool
}

// readAppleArchive reads an ipa or a zipped .app
func readAppleArchive(files []*zip.File, info *BundleInfo) (root string, err error) {
	root, mac, err := appleRoot(files)
	if err != nil {
		return "", err
	}
	b := appleBundle{files: files, root: root, mac: mac}
	content, err := b.readInfo(info)
	if err != nil {
		return "", err
	}
	b.adaptMac(info.Apple)
	info.Icon = b.icon(content)
	b.readProfile(info)
	b.listComponents(info.Apple)
	return root, nil
}

func (b appleBundle) readInfo(info *BundleInfo) (plistDict, error) {
	data, err := readZipFile(findFile(b.files, b.root+"Info.plist"), 16<<20)
	if err != nil {
		return nil, err
	}
	return readInfoPlist(data, info)
}

// adaptMac sets what a Mac app does not say in its Info.plist
func (b appleBundle) adaptMac(apple *AppleInfo) {
	if !b.mac {
		return
	}
	if apple.Platform == "" {
		apple.Platform = "macosx"
	}
	prefixMacExecutable(apple)
}

// prefixMacExecutable makes the executable relative to the .app
func prefixMacExecutable(apple *AppleInfo) {
	if apple.Executable != "" {
		apple.Executable = "Contents/MacOS/" + apple.Executable
	}
}

// resourceDir is the folder of the icons
func (b appleBundle) resourceDir() string {
	if b.mac {
		return b.root + "Resources/"
	}
	return b.root
}

func (b appleBundle) profileName() string {
	if b.mac {
		return "embedded.provisionprofile"
	}
	return "embedded.mobileprovision"
}

// readProfile reads the provisioning profile; a profile that cannot be read
// is a warning, not a failure
func (b appleBundle) readProfile(info *BundleInfo) {
	f := findFile(b.files, b.root+b.profileName())
	if f == nil {
		return
	}
	profile, certs, err := readProfileFile(f)
	if err != nil {
		info.Warnings = append(info.Warnings, err.Error())
		return
	}
	info.Apple.Profile = profile
	if len(certs) > 0 {
		info.Signing = &Signing{Schemes: []string{"profile"}, Certificates: certs}
	}
}

func readProfileFile(f *zip.File) (*ProvisioningProfile, []Certificate, error) {
	data, err := readZipFile(f, 4<<20)
	if err != nil {
		return nil, nil, err
	}
	return readProvisioning(data)
}

// listComponents lists the frameworks and the extensions of the app, each once
func (b appleBundle) listComponents(apple *AppleInfo) {
	seen := map[string]bool{}
	for _, f := range b.files {
		rel, found := strings.CutPrefix(f.Name, b.root)
		if !found {
			continue
		}
		dir, name, ok := componentOf(rel)
		if !ok || seen[dir+name] {
			continue
		}
		seen[dir+name] = true
		addComponent(apple, dir, name)
	}
}

// componentOf is the folder (PlugIns/, Frameworks/, Extensions/) and the
// name of the component a path inside the app belongs to
func componentOf(rel string) (dir, name string, ok bool) {
	for _, dir := range []string{"PlugIns/", "Frameworks/", "Extensions/"} {
		if rest, found := strings.CutPrefix(rel, dir); found {
			name, _, _ = strings.Cut(rest, "/")
			return dir, name, name != ""
		}
	}
	return "", "", false
}

func addComponent(apple *AppleInfo, dir, name string) {
	switch {
	case dir == "Frameworks/":
		apple.Frameworks = append(apple.Frameworks, name)
	case strings.HasSuffix(name, ".appex") || strings.HasSuffix(name, ".xctest"):
		apple.Extensions = append(apple.Extensions, name)
	}
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
	if err == nil && mac {
		prefixMacExecutable(info.Apple)
	}
	return mac, err
}
