package config

type Connection struct {
	Type string `mapstructure:"type"`
	IP   string `mapstructure:"ip"`
}

type Device struct {
	ID         string     `mapstructure:"id"`
	IsTablet   bool       `mapstructure:"table"`
	Name       string     `mapstructure:"name"`
	Manager    string     `mapstructure:"manager"`
	PIN        string     `mapstructure:"pin"`
	Connection Connection `mapstructure:"connection"`
}

type Service struct {
	Autodetect bool     `mapstructure:"autodetect_ip"`
	IP         string   `mapstructure:"ip"`
	Port       uint16   `mapstructure:"port"`
	Devices    []Device `mapstructure:"devices"`
}
