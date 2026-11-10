package app

import "time"

// BundleInfo is everything the master reads out of an uploaded apk, ipa or
// zipped app. It is stored with the bundle and shown on its details page.
type BundleInfo struct {
	// apk, ipa, zip or app
	Format      string `json:"format"`
	DisplayName string `json:"displayName,omitempty"`
	Identifier  string `json:"identifier,omitempty"`
	// the version people see: versionName or CFBundleShortVersionString
	Version string `json:"version,omitempty"`
	// the build number: versionCode or CFBundleVersion
	Build string `json:"build,omitempty"`
	// path of the icon inside the archive, served by the icon endpoint
	Icon     string       `json:"icon,omitempty"`
	Engine   *Engine      `json:"engine,omitempty"`
	Android  *AndroidInfo `json:"android,omitempty"`
	Apple    *AppleInfo   `json:"apple,omitempty"`
	Signing  *Signing     `json:"signing,omitempty"`
	Contents *Contents    `json:"contents,omitempty"`
	// things that could not be read; the rest of the info is still valid
	Warnings []string  `json:"warnings,omitempty"`
	ReadAt   time.Time `json:"readAt"`
}

type Engine struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	// IL2CPP or Mono for Unity
	Backend string `json:"backend,omitempty"`
}

type AndroidPermission struct {
	Name   string `json:"name"`
	MaxSdk string `json:"maxSdk,omitempty"`
}

type AndroidFeature struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

type AndroidInfo struct {
	PackageName          string              `json:"packageName"`
	VersionCode          string              `json:"versionCode,omitempty"`
	VersionName          string              `json:"versionName,omitempty"`
	MinSdk               string              `json:"minSdk,omitempty"`
	TargetSdk            string              `json:"targetSdk,omitempty"`
	MaxSdk               string              `json:"maxSdk,omitempty"`
	CompileSdk           string              `json:"compileSdk,omitempty"`
	CompileSdkCodename   string              `json:"compileSdkCodename,omitempty"`
	PlatformBuildVersion string              `json:"platformBuildVersion,omitempty"`
	Label                string              `json:"label,omitempty"`
	LaunchActivity       string              `json:"launchActivity,omitempty"`
	Permissions          []AndroidPermission `json:"permissions,omitempty"`
	Features             []AndroidFeature    `json:"features,omitempty"`
	ABIs                 []string            `json:"abis,omitempty"`
	Locales              []string            `json:"locales,omitempty"`
	Densities            []string            `json:"densities,omitempty"`
	Screens              []string            `json:"screens,omitempty"`
	GLES                 string              `json:"gles,omitempty"`
	Debuggable           bool                `json:"debuggable"`
	TestOnly             bool                `json:"testOnly"`
	Leanback             bool                `json:"leanback"`
}

type UsageDescription struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

type AppleInfo struct {
	// iphoneos, macosx, iphonesimulator ...
	Platform             string               `json:"platform,omitempty"`
	MinOS                string               `json:"minOs,omitempty"`
	SDK                  string               `json:"sdk,omitempty"`
	Xcode                string               `json:"xcode,omitempty"`
	XcodeBuild           string               `json:"xcodeBuild,omitempty"`
	BuildMachineOS       string               `json:"buildMachineOs,omitempty"`
	BundleName           string               `json:"bundleName,omitempty"`
	Executable           string               `json:"executable,omitempty"`
	DeviceFamily         []string             `json:"deviceFamily,omitempty"`
	RequiredCapabilities []string             `json:"requiredCapabilities,omitempty"`
	Orientations         []string             `json:"orientations,omitempty"`
	BackgroundModes      []string             `json:"backgroundModes,omitempty"`
	URLSchemes           []string             `json:"urlSchemes,omitempty"`
	UsageDescriptions    []UsageDescription   `json:"usageDescriptions,omitempty"`
	Extensions           []string             `json:"extensions,omitempty"`
	Frameworks           []string             `json:"frameworks,omitempty"`
	FileSharing          bool                 `json:"fileSharing"`
	Profile              *ProvisioningProfile `json:"profile,omitempty"`
}

type ProvisioningProfile struct {
	Name      string `json:"name"`
	UUID      string `json:"uuid,omitempty"`
	AppIDName string `json:"appIdName,omitempty"`
	TeamName  string `json:"teamName,omitempty"`
	TeamID    string `json:"teamId,omitempty"`
	// development, ad-hoc, enterprise or app-store
	Kind         string                 `json:"kind"`
	CreatedAt    time.Time              `json:"createdAt"`
	ExpiresAt    time.Time              `json:"expiresAt"`
	AllDevices   bool                   `json:"allDevices"`
	Devices      []string               `json:"devices,omitempty"`
	Entitlements map[string]interface{} `json:"entitlements,omitempty"`
}

type Certificate struct {
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	Serial    string    `json:"serial"`
	NotBefore time.Time `json:"notBefore"`
	NotAfter  time.Time `json:"notAfter"`
	SHA1      string    `json:"sha1"`
	SHA256    string    `json:"sha256"`
	// signed with the Android debug key
	Debug bool `json:"debug"`
}

type Signing struct {
	// v1, v2, v3, v3.1 for apks; profile for apple apps
	Schemes      []string      `json:"schemes,omitempty"`
	Certificates []Certificate `json:"certificates,omitempty"`
}

type SizeEntry struct {
	Path       string `json:"path"`
	Files      int    `json:"files,omitempty"`
	Size       int64  `json:"size"`
	Compressed int64  `json:"compressed"`
}

type Contents struct {
	Files      int         `json:"files"`
	Size       int64       `json:"size"`
	Compressed int64       `json:"compressed"`
	Groups     []SizeEntry `json:"groups,omitempty"`
	Largest    []SizeEntry `json:"largest,omitempty"`
}
