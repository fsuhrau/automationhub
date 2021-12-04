package config

/*
IOSSim        bool `mapstructure:"iossim"`
MacOS         bool `mapstructure:"macos"`
IOSDevice     bool `mapstructure:"iosdevice"`
AndroidDevice bool `mapstructure:"androiddevice"`
*/

type Connection struct {
	Type string `mapstructure:"type"`
	IP   string `mapstructure:"ip"`
}

type Manager struct {
	Enabled         bool     `mapstructure:"enabled"`
	UseOSScreenshot bool     `mapstructure:"use_os_screenshot"`
	Devices         []Device `mapstructure:"devices"`
}

type Device struct {
	ID         string     `mapstructure:"id"`
	IsTablet   bool       `mapstructure:"tablet"`
	Name       string     `mapstructure:"name"`
	PIN        string     `mapstructure:"pin"`
	Connection Connection `mapstructure:"connection"`
}

type Hook struct {
	Provider string `mapstructure:"provider"`
	Url      string `mapstructure:"url"`
	Token    string `mapstructure:"token"`
	Username string `mapstructure:"username"`
	Channel  string `mapstructure:"channel"`
}

type Service struct {
	Autodetect    bool               `mapstructure:"autodetect_ip"`
	AutoDiscovery bool               `mapstructure:"auto_discovery"`
	HostIP        string             `mapstructure:"host_ip"`
	ServerURL     string             `mapstructure:"server_url"`
	Port          uint16             `mapstructure:"port"`
	DeviceManager map[string]Manager `mapstructure:"managers"`
	Hooks         []Hook             `mapstructure:"hooks"`
}
