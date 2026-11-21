package app

import (
	"archive/zip"
	"sort"
	"strings"
)

const (
	maxGroups  = 20
	maxLargest = 15
)

// isGroupedDir says whether a folder holds distinct parts of a build, so its
// children are shown on their own (lib/arm64-v8a, assets/aa, Frameworks/X)
func isGroupedDir(name string) bool {
	switch name {
	case "lib", "assets", "res", "Frameworks", "PlugIns", "Data", "Resources", "Contents", "META-INF":
		return true
	}
	return false
}

// groupOf is the part of a build a file belongs to
func groupOf(rel string) string {
	parts := strings.Split(rel, "/")
	switch {
	case len(parts) == 1:
		return parts[0]
	case !isGroupedDir(parts[0]) || len(parts) <= 2:
		return parts[0] + "/"
	case parts[0] == "assets" && parts[1] == "bin" && len(parts) > 3:
		// assets/bin/Data is one part of a Unity build, not "assets/bin"
		return strings.Join(parts[:3], "/") + "/"
	}
	return parts[0] + "/" + parts[1] + "/"
}

func (e *SizeEntry) addFile(f SizeEntry) {
	e.Files++
	e.Size += f.Size
	e.Compressed += f.Compressed
}

func (e *SizeEntry) merge(o SizeEntry) {
	e.Files += o.Files
	e.Size += o.Size
	e.Compressed += o.Compressed
}

// readContents sums up the files of an archive below root, by part and the
// largest files
func readContents(files []*zip.File, root string) *Contents {
	c := &Contents{}
	groups := map[string]*SizeEntry{}
	var all []SizeEntry
	for _, f := range files {
		rel, ok := strings.CutPrefix(f.Name, root)
		if f.FileInfo().IsDir() || !ok || rel == "" {
			continue
		}
		file := SizeEntry{Path: rel, Size: int64(f.UncompressedSize64), Compressed: int64(f.CompressedSize64)}
		c.Files++
		c.Size += file.Size
		c.Compressed += file.Compressed
		groupEntry(groups, groupOf(rel)).addFile(file)
		all = append(all, file)
	}
	c.Groups = topGroups(groups)
	c.Largest = largestFiles(all)
	return c
}

func groupEntry(groups map[string]*SizeEntry, key string) *SizeEntry {
	g, ok := groups[key]
	if !ok {
		g = &SizeEntry{Path: key}
		groups[key] = g
	}
	return g
}

// topGroups sorts the groups by size and sums up the smallest ones as "other"
// when there are too many
func topGroups(groups map[string]*SizeEntry) []SizeEntry {
	var list []SizeEntry
	for _, g := range groups {
		list = append(list, *g)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Size > list[j].Size })
	if len(list) <= maxGroups {
		return list
	}
	rest := SizeEntry{Path: "other"}
	for _, g := range list[maxGroups-1:] {
		rest.merge(g)
	}
	return append(list[:maxGroups-1], rest)
}

func largestFiles(all []SizeEntry) []SizeEntry {
	sort.Slice(all, func(i, j int) bool { return all[i].Size > all[j].Size })
	return all[:min(len(all), maxLargest)]
}
