package app

import (
	"archive/zip"
	"slices"
	"sort"
	"strings"
)

// apkIcon picks the biggest png icon. Adaptive icons are xml, so a png with
// the same name in a density folder is taken instead.
func apkIcon(files []*zip.File, icons map[int]string) string {
	paths := iconPathsByDensity(icons)
	for _, path := range paths {
		if isRasterIcon(path) && findFile(files, path) != nil {
			return path
		}
	}
	names := make([]string, len(paths))
	for i, path := range paths {
		base := path[strings.LastIndex(path, "/")+1:]
		names[i] = strings.TrimSuffix(base, ".xml")
	}
	return bestDensityIcon(files, names)
}

// iconPathsByDensity lists the icon paths, the biggest density first
func iconPathsByDensity(icons map[int]string) []string {
	densities := make([]int, 0, len(icons))
	for d := range icons {
		densities = append(densities, d)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(densities)))
	paths := make([]string, len(densities))
	for i, d := range densities {
		paths[i] = icons[d]
	}
	return paths
}

func isRasterIcon(path string) bool {
	return strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".webp")
}

// bestDensityIcon finds the png or webp of the highest density folder that is
// named like one of the names
func bestDensityIcon(files []*zip.File, names []string) string {
	best, bestRank := "", -1
	for _, f := range files {
		stem, ok := densityIconStem(f.Name)
		if !ok || !slices.Contains(names, stem) {
			continue
		}
		if rank := densityRank(f.Name); rank > bestRank {
			best, bestRank = f.Name, rank
		}
	}
	return best
}

// densityIconStem is the name without extension of a png or webp in a mipmap
// or drawable folder
func densityIconStem(path string) (string, bool) {
	if !strings.HasPrefix(path, "res/mipmap-") && !strings.HasPrefix(path, "res/drawable-") {
		return "", false
	}
	base := path[strings.LastIndex(path, "/")+1:]
	i := strings.LastIndex(base, ".")
	if i <= 0 || (base[i:] != ".png" && base[i:] != ".webp") {
		return "", false
	}
	return base[:i], true
}

func densityRank(path string) int {
	for i, d := range []string{"xxxhdpi", "xxhdpi", "xhdpi", "hdpi", "mdpi", "ldpi"} {
		if strings.Contains(path, "-"+d) {
			return 10 - i
		}
	}
	return 0
}
