package config

type Manager struct {
	Enabled         bool     `mapstructure:"enabled"`
	UseOSScreenshot bool     `mapstructure:"use_os_screenshot"`
}

type Hook struct {
	Provider string `mapstructure:"provider"`
	Url      string `mapstructure:"url"`
	Token    string `mapstructure:"token"`
	Username string `mapstructure:"username"`
	Channel  string `mapstructure:"channel"`
}

type Service struct {
	// Port          uint16             `mapstructure:"port"`
	Autodetect    bool               `mapstructure:"autodetect_ip"`
	AutoDiscovery bool               `mapstructure:"auto_discovery"`
	HostIP        string             `mapstructure:"host_ip"`
	ServerURL     string             `mapstructure:"server_url"`
	DeviceManager map[string]Manager `mapstructure:"managers"`
	Hooks         []Hook             `mapstructure:"hooks"`
}
