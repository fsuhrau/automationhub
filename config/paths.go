package config

import "path/filepath"

// Paths are the folders the hub keeps files in.
type Paths struct {
	// AppStore holds the uploaded app bundles.
	AppStore string `yaml:"app_store,omitempty" mapstructure:"app_store"`
	// Bundles holds the bundles a node received from the master.
	Bundles string `yaml:"bundles,omitempty" mapstructure:"bundles"`
	// TestData holds the screenshots and recordings of test runs.
	TestData string `yaml:"test_data,omitempty" mapstructure:"test_data"`
}

// DefaultPaths are the folders below the .automationhub folder of home.
func DefaultPaths(home string) Paths {
	base := filepath.Join(home, ".automationhub")
	return Paths{
		AppStore: filepath.Join(base, "store"),
		Bundles:  filepath.Join(base, "bundles"),
		TestData: filepath.Join(base, "data"),
	}
}

// WithDefaults returns the paths with every empty one taken from defaults.
func (p Paths) WithDefaults(defaults Paths) Paths {
	p.AppStore = firstNonEmpty(p.AppStore, defaults.AppStore)
	p.Bundles = firstNonEmpty(p.Bundles, defaults.Bundles)
	p.TestData = firstNonEmpty(p.TestData, defaults.TestData)
	return p
}

// All lists the folders, for creating them.
func (p Paths) All() []string {
	return []string{p.AppStore, p.Bundles, p.TestData}
}

func firstNonEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
