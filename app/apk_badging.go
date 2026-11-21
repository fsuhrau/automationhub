package app

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var iconDensityRegex = regexp.MustCompile(`^application-icon-(\d+)$`)

// badgingParser collects the AndroidInfo while the lines of the aapt output
// are fed to it
type badgingParser struct {
	info     *AndroidInfo
	icons    map[int]string
	features featureSet
	handlers map[string]func(badgingLine)
}

// parseBadging reads the output of "aapt dump badging"
func parseBadging(output []byte) (*AndroidInfo, map[int]string) {
	p := newBadgingParser()
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for scanner.Scan() {
		p.parseLine(strings.TrimSpace(scanner.Text()))
	}
	p.info.Features = p.features.list()
	p.info.Permissions = uniquePermissions(p.info.Permissions)
	return p.info, p.icons
}

func newBadgingParser() *badgingParser {
	p := &badgingParser{info: &AndroidInfo{}, icons: map[int]string{}}
	p.handlers = map[string]func(badgingLine){
		"package":                   p.packageLine,
		"sdkVersion":                func(l badgingLine) { p.info.MinSdk = l.value() },
		"minSdkVersion":             func(l badgingLine) { p.info.MinSdk = l.value() },
		"targetSdkVersion":          func(l badgingLine) { p.info.TargetSdk = l.value() },
		"maxSdkVersion":             func(l badgingLine) { p.info.MaxSdk = l.value() },
		"application-label":         func(l badgingLine) { p.info.Label = l.value() },
		"application":               p.applicationLine,
		"launchable-activity":       p.launchableLine,
		"uses-permission":           p.permissionLine,
		"uses-permission-sdk-23":    p.permissionLine,
		"uses-permission-sdk-m":     p.permissionLine,
		"uses-feature":              func(l badgingLine) { p.features.add(l.attrs["name"], l.attrs["required"] != "false") },
		"uses-feature-not-required": func(l badgingLine) { p.features.add(l.attrs["name"], false) },
		"uses-gl-es":                func(l badgingLine) { p.info.GLES = glesVersion(l.value()) },
		"native-code":               p.nativeCodeLine,
		"alt-native-code":           p.nativeCodeLine,
		"locales":                   p.localesLine,
		"densities":                 func(l badgingLine) { p.info.Densities = l.list() },
		"supports-screens":          func(l badgingLine) { p.info.Screens = l.list() },
	}
	return p
}

func (p *badgingParser) parseLine(text string) {
	key, rest, found := strings.Cut(text, ":")
	if !found {
		p.flagLine(text)
		return
	}
	line := newBadgingLine(key, rest)
	if handle, ok := p.handlers[key]; ok {
		handle(line)
		return
	}
	p.iconLine(line)
}

// flagLine handles the lines without a colon
func (p *badgingParser) flagLine(text string) {
	switch {
	case text == "application-debuggable":
		p.info.Debuggable = true
	case text == "leanback-launcher":
		p.info.Leanback = true
	case strings.HasPrefix(text, "testOnly="):
		p.info.TestOnly = true
	}
}

func (p *badgingParser) packageLine(l badgingLine) {
	p.info.PackageName = l.attrs["name"]
	p.info.VersionCode = l.attrs["versionCode"]
	p.info.VersionName = l.attrs["versionName"]
	p.info.CompileSdk = l.attrs["compileSdkVersion"]
	p.info.CompileSdkCodename = l.attrs["compileSdkVersionCodename"]
	p.info.PlatformBuildVersion = l.attrs["platformBuildVersionName"]
}

func (p *badgingParser) applicationLine(l badgingLine) {
	if p.info.Label == "" {
		p.info.Label = l.attrs["label"]
	}
	if _, ok := p.icons[0]; !ok && l.attrs["icon"] != "" {
		p.icons[0] = l.attrs["icon"]
	}
}

func (p *badgingParser) launchableLine(l badgingLine) {
	if p.info.LaunchActivity == "" {
		p.info.LaunchActivity = l.attrs["name"]
	}
}

func (p *badgingParser) permissionLine(l badgingLine) {
	if name := l.attrs["name"]; name != "" {
		p.info.Permissions = append(p.info.Permissions, AndroidPermission{Name: name, MaxSdk: l.attrs["maxSdkVersion"]})
	}
}

func (p *badgingParser) nativeCodeLine(l badgingLine) {
	p.info.ABIs = append(p.info.ABIs, l.list()...)
}

func (p *badgingParser) localesLine(l badgingLine) {
	for _, locale := range l.list() {
		if locale != "--_--" {
			p.info.Locales = append(p.info.Locales, locale)
		}
	}
}

// iconLine handles application-icon-<density>, the only line kind that is
// not a fixed key
func (p *badgingParser) iconLine(l badgingLine) {
	if m := iconDensityRegex.FindStringSubmatch(l.key); m != nil {
		density, _ := strconv.Atoi(m[1])
		p.icons[density] = l.value()
	}
}

// featureSet keeps the features in the order they first appear; a feature is
// required when any of its lines says so
type featureSet struct {
	order    []string
	required map[string]bool
}

func (f *featureSet) add(name string, required bool) {
	if name == "" {
		return
	}
	if f.required == nil {
		f.required = map[string]bool{}
	}
	if _, ok := f.required[name]; !ok {
		f.order = append(f.order, name)
	}
	f.required[name] = f.required[name] || required
}

func (f *featureSet) list() []AndroidFeature {
	var out []AndroidFeature
	for _, name := range f.order {
		out = append(out, AndroidFeature{Name: name, Required: f.required[name]})
	}
	return out
}

func uniquePermissions(in []AndroidPermission) []AndroidPermission {
	seen := map[string]bool{}
	var out []AndroidPermission
	for _, p := range in {
		if !seen[p.Name] {
			seen[p.Name] = true
			out = append(out, p)
		}
	}
	return out
}

// glesVersion turns aapt's '0x00030000' into 3.0
func glesVersion(hex string) string {
	v, err := strconv.ParseUint(strings.TrimPrefix(hex, "0x"), 16, 32)
	if err != nil {
		return hex
	}
	return fmt.Sprintf("%d.%d", v>>16, v&0xffff)
}
