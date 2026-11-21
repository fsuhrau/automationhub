package app

import (
	"archive/zip"
	"errors"
	"strings"
)

type appCandidate struct {
	path string
	mac  bool
}

// appleRoot finds the folder of the app inside an ipa or a zipped .app:
// Payload/X.app/ for an ipa, X.app/ or X.app/Contents/ for a zip. The app
// with the shortest path wins, so plug-ins inside it are skipped.
func appleRoot(files []*zip.File) (root string, mac bool, err error) {
	var best appCandidate
	for _, f := range files {
		c, ok := appCandidateOf(f.Name)
		if ok && (best.path == "" || len(c.path) < len(best.path)) {
			best = c
		}
	}
	if best.path == "" {
		return "", false, errors.New("no .app with an Info.plist in the archive")
	}
	return best.path, best.mac, nil
}

// appCandidateOf says whether the file is the Info.plist of an app and where
// that app is
func appCandidateOf(name string) (appCandidate, bool) {
	switch {
	case strings.HasPrefix(name, "__MACOSX/"):
		return appCandidate{}, false
	case strings.HasSuffix(name, ".app/Contents/Info.plist"):
		return appCandidate{path: strings.TrimSuffix(name, "Info.plist"), mac: true}, true
	case strings.HasSuffix(name, ".app/Info.plist"):
		return appCandidate{path: strings.TrimSuffix(name, "Info.plist")}, true
	}
	return appCandidate{}, false
}
