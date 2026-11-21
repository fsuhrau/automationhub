package app

import (
	"path/filepath"
	"strings"
)

// iconNames lists the icon file names an Info.plist names
func iconNames(content plistDict) []string {
	var names []string
	for _, key := range []string{"CFBundleIcons", "CFBundleIcons~ipad"} {
		names = append(names, primaryIconNames(content[key])...)
	}
	names = append(names, plistStrings(content, "CFBundleIconFiles")...)
	if n := plistString(content, "CFBundleIconFile"); n != "" {
		names = append(names, n)
	}
	return names
}

// primaryIconNames reads the CFBundlePrimaryIcon of a CFBundleIcons dict
func primaryIconNames(icons interface{}) []string {
	dict, _ := icons.(plistDict)
	primary, ok := dict["CFBundlePrimaryIcon"].(plistDict)
	if !ok {
		return nil
	}
	names := plistStrings(primary, "CFBundleIconFiles")
	if n := plistString(primary, "CFBundleIconName"); n != "" {
		names = append(names, n)
	}
	return names
}

// icon picks the largest icon file of the app in the archive
func (b appleBundle) icon(content plistDict) string {
	names := iconNames(content)
	if len(names) == 0 {
		return ""
	}
	dir := b.resourceDir()
	best := ""
	var bestSize uint64
	for _, f := range b.files {
		base, ok := fileInDir(f.Name, dir)
		if ok && isIconFile(base, names) && f.UncompressedSize64 > bestSize {
			best, bestSize = f.Name, f.UncompressedSize64
		}
	}
	return best
}

// fileInDir returns the name of a file that sits directly in dir
func fileInDir(path, dir string) (string, bool) {
	base, found := strings.CutPrefix(path, dir)
	if !found || strings.Contains(base, "/") {
		return "", false
	}
	return base, true
}

// isIconFile says whether a png or icns is one of the named icons, in any
// scale (name@2x, name~ipad)
func isIconFile(base string, names []string) bool {
	ext := filepath.Ext(base)
	if lower := strings.ToLower(ext); lower != ".png" && lower != ".icns" {
		return false
	}
	stem := strings.TrimSuffix(base, ext)
	for _, n := range names {
		n = strings.TrimSuffix(n, filepath.Ext(n))
		if stem == n || strings.HasPrefix(stem, n+"@") || strings.HasPrefix(stem, n+"~") {
			return true
		}
	}
	return false
}
