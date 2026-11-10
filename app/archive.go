package app

import (
	"archive/zip"
	"io"
	"regexp"
	"sort"
	"strings"
)

const (
	maxGroups  = 20
	maxLargest = 15
)

var unityVersionRegex = regexp.MustCompile(`(?:20\d\d|6\d{3}|[45])\.\d+\.\d+[abfpx]\d+`)

// groupedDirs are folders that hold distinct parts of a build, so their
// children are shown on their own (lib/arm64-v8a, assets/aa, Frameworks/X)
var groupedDirs = map[string]bool{
	"lib": true, "assets": true, "res": true, "Frameworks": true, "PlugIns": true,
	"Data": true, "Resources": true, "Contents": true, "META-INF": true,
}

// groupOf is the part of a build a file belongs to
func groupOf(rel string) string {
	parts := strings.Split(rel, "/")
	if len(parts) == 1 {
		return parts[0]
	}
	if groupedDirs[parts[0]] && len(parts) > 2 {
		// assets/bin/Data is one part of a Unity build, not "assets/bin"
		if parts[0] == "assets" && parts[1] == "bin" && len(parts) > 3 {
			return strings.Join(parts[:3], "/") + "/"
		}
		return parts[0] + "/" + parts[1] + "/"
	}
	return parts[0] + "/"
}

// readContents sums up the files of an archive below root, by part and the
// largest files
func readContents(files []*zip.File, root string) *Contents {
	c := &Contents{}
	groups := map[string]*SizeEntry{}
	var all []SizeEntry
	for _, f := range files {
		if f.FileInfo().IsDir() || !strings.HasPrefix(f.Name, root) {
			continue
		}
		rel := strings.TrimPrefix(f.Name, root)
		if rel == "" {
			continue
		}
		size, comp := int64(f.UncompressedSize64), int64(f.CompressedSize64)
		c.Files++
		c.Size += size
		c.Compressed += comp
		key := groupOf(rel)
		g, ok := groups[key]
		if !ok {
			g = &SizeEntry{Path: key}
			groups[key] = g
		}
		g.Files++
		g.Size += size
		g.Compressed += comp
		all = append(all, SizeEntry{Path: rel, Size: size, Compressed: comp})
	}
	for _, g := range groups {
		c.Groups = append(c.Groups, *g)
	}
	sort.Slice(c.Groups, func(i, j int) bool { return c.Groups[i].Size > c.Groups[j].Size })
	if len(c.Groups) > maxGroups {
		rest := SizeEntry{Path: "other"}
		for _, g := range c.Groups[maxGroups-1:] {
			rest.Files += g.Files
			rest.Size += g.Size
			rest.Compressed += g.Compressed
		}
		c.Groups = append(c.Groups[:maxGroups-1], rest)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Size > all[j].Size })
	if len(all) > maxLargest {
		all = all[:maxLargest]
	}
	c.Largest = all
	return c
}

// detectEngine recognises the engine or framework a build was made with
func detectEngine(files []*zip.File, root string) *Engine {
	var engine *Engine
	var managers *zip.File
	il2cpp, mono := false, false
	for _, f := range files {
		if !strings.HasPrefix(f.Name, root) {
			continue
		}
		name := strings.TrimPrefix(f.Name, root)
		base := name[strings.LastIndex(name, "/")+1:]
		switch {
		case base == "libunity.so" || strings.HasPrefix(name, "Frameworks/UnityFramework.framework/") ||
			base == "UnityPlayer.dylib" || base == "UnityPlayer.dll":
			engine = &Engine{Name: "Unity"}
		case base == "globalgamemanagers" && (managers == nil || len(name) < len(managers.Name)):
			managers = f
		case base == "libil2cpp.so" || base == "global-metadata.dat" || base == "GameAssembly.dylib":
			il2cpp = true
		case base == "libmonobdwgc-2.0.so" || base == "libmonosgen-2.0.so" || strings.HasSuffix(name, "Managed/mscorlib.dll") && !il2cpp:
			mono = true
		}
		if engine == nil {
			switch {
			case base == "libUE4.so" || base == "libUnreal.so":
				engine = &Engine{Name: "Unreal Engine"}
			case base == "libflutter.so" || strings.HasPrefix(name, "Frameworks/Flutter.framework/"):
				engine = &Engine{Name: "Flutter"}
			case base == "libreactnativejni.so" || base == "index.android.bundle" || base == "main.jsbundle":
				engine = &Engine{Name: "React Native"}
			case base == "libgodot_android.so":
				engine = &Engine{Name: "Godot"}
			case base == "libcocos2dcpp.so" || base == "libcocos.so":
				engine = &Engine{Name: "Cocos"}
			}
		}
	}
	if engine == nil && managers != nil {
		engine = &Engine{Name: "Unity"}
	}
	if engine == nil {
		return nil
	}
	if engine.Name == "Unity" {
		if managers != nil {
			engine.Version = unityVersion(managers)
		}
		if il2cpp {
			engine.Backend = "IL2CPP"
		} else if mono {
			engine.Backend = "Mono"
		}
	}
	return engine
}

// unityVersion reads the editor version from the header of a serialized file
func unityVersion(f *zip.File) string {
	rc, err := f.Open()
	if err != nil {
		return ""
	}
	defer rc.Close()
	head := make([]byte, 256)
	n, _ := io.ReadFull(rc, head)
	return unityVersionRegex.FindString(string(head[:n]))
}

func findFile(files []*zip.File, name string) *zip.File {
	for _, f := range files {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func readZipFile(f *zip.File, limit int64) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(io.LimitReader(rc, limit))
}
