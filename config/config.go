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
	IsTablet   bool       `mapstructure:"table"`
	Name       string     `mapstructure:"name"`
	PIN        string     `mapstructure:"pin"`
	Connection Connection `mapstructure:"connection"`
}

type Service struct {
	Autodetect    bool               `mapstructure:"autodetect_ip"`
	HostIP        string             `mapstructure:"host_ip"`
	ServerURL     string             `mapstructure:"server_url"`
	Port          uint16             `mapstructure:"port"`
	DeviceManager map[string]Manager `mapstructure:"managers"`
}
