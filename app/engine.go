package app

import (
	"archive/zip"
	"io"
	"regexp"
	"strings"
)

var unityVersionRegex = regexp.MustCompile(`(?:20\d\d|6\d{3}|[45])\.\d+\.\d+[abfpx]\d+`)

// engineScan collects the hints of the engine while it goes over the files
type engineScan struct {
	engine       *Engine
	managers     *zip.File
	il2cpp, mono bool
}

// detectEngine recognises the engine or framework a build was made with
func detectEngine(files []*zip.File, root string) *Engine {
	s := &engineScan{}
	for _, f := range files {
		if name, ok := strings.CutPrefix(f.Name, root); ok {
			s.visit(f, name)
		}
	}
	return s.result()
}

func (s *engineScan) visit(f *zip.File, name string) {
	base := name[strings.LastIndex(name, "/")+1:]
	switch {
	case isUnityFile(name, base):
		s.engine = &Engine{Name: "Unity"}
	case base == "globalgamemanagers" && (s.managers == nil || len(name) < len(s.managers.Name)):
		s.managers = f
	case base == "libil2cpp.so" || base == "global-metadata.dat" || base == "GameAssembly.dylib":
		s.il2cpp = true
	case isMonoFile(name, base, s.il2cpp):
		s.mono = true
	}
	if s.engine == nil {
		s.engine = otherEngine(name, base)
	}
}

func isUnityFile(name, base string) bool {
	return base == "libunity.so" || strings.HasPrefix(name, "Frameworks/UnityFramework.framework/") ||
		base == "UnityPlayer.dylib" || base == "UnityPlayer.dll"
}

// isMonoFile recognises the Mono runtime; the managed corlib only counts
// when the build has no IL2CPP, as IL2CPP builds ship it as well
func isMonoFile(name, base string, il2cpp bool) bool {
	return base == "libmonobdwgc-2.0.so" || base == "libmonosgen-2.0.so" ||
		(!il2cpp && strings.HasSuffix(name, "Managed/mscorlib.dll"))
}

// otherEngine recognises the engines that are not Unity
func otherEngine(name, base string) *Engine {
	switch {
	case base == "libUE4.so" || base == "libUnreal.so":
		return &Engine{Name: "Unreal Engine"}
	case base == "libflutter.so" || strings.HasPrefix(name, "Frameworks/Flutter.framework/"):
		return &Engine{Name: "Flutter"}
	case base == "libreactnativejni.so" || base == "index.android.bundle" || base == "main.jsbundle":
		return &Engine{Name: "React Native"}
	case base == "libgodot_android.so":
		return &Engine{Name: "Godot"}
	case base == "libcocos2dcpp.so" || base == "libcocos.so":
		return &Engine{Name: "Cocos"}
	}
	return nil
}

// result is the engine found; a Unity build without its player library is
// still recognised by the serialized managers file
func (s *engineScan) result() *Engine {
	if s.engine == nil && s.managers != nil {
		s.engine = &Engine{Name: "Unity"}
	}
	if s.engine != nil && s.engine.Name == "Unity" {
		s.fillUnity()
	}
	return s.engine
}

func (s *engineScan) fillUnity() {
	if s.managers != nil {
		s.engine.Version = unityVersion(s.managers)
	}
	switch {
	case s.il2cpp:
		s.engine.Backend = "IL2CPP"
	case s.mono:
		s.engine.Backend = "Mono"
	}
}

// unityVersion reads the editor version from the header of a serialized file
func unityVersion(f *zip.File) string {
	rc, err := f.Open()
	if err != nil {
		return ""
	}
	defer closeQuietly(rc)
	head := make([]byte, 256)
	n, _ := io.ReadFull(rc, head)
	return unityVersionRegex.FindString(string(head[:n]))
}
