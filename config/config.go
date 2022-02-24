package config

type Manager struct {
	Enabled         bool `mapstructure:"enabled"`
	UseOSScreenshot bool `mapstructure:"use_os_screenshot"`
}

type Hook struct {
	Provider string `mapstructure:"provider"`
	Url      string `mapstructure:"url"`
	Token    string `mapstructure:"token"`
	Username string `mapstructure:"username"`
	Channel  string `mapstructure:"channel"`
}

type Token struct {
}

type Auth struct {
	Token  *Token  `mapstructure:"token"`
	OAuth2 *OAuth2 `mapstructure:"oauth2"`
	Github *Github `mapstrucutre:"github"`
}

func (a *Auth) AuthenticationRequired() bool {
	return a.OAuth2 != nil || a.Github != nil || a.Token != nil
}

type OAuth2 struct {
	AuthUrl     string   `mapstructure:"auth_url"`
	TokenUrl    string   `mapstructure:"token_url"`
	RedirectUrl string   `mapstructure:"redirect_url"`
	UserUrl     string   `mapstructure:"user_url"`
	Credentials string   `mapstructure:"credentials"`
	Scopes      []string `mapstructure:"scopes"`
	Secret      string   `mapstructure:"secret"`
}

type Github struct {
	RedirectUrl string   `mapstructure:"redirect_url"`
	Credentials string   `mapstructure:"credentials"`
	Scopes      []string `mapstructure:"scopes"`
	Secret      string   `mapstructure:"secret"`
}

type Service struct {
	Port          uint16             `mapstructure:"port"`
	Autodetect    bool               `mapstructure:"autodetect_ip"`
	AutoDiscovery bool               `mapstructure:"auto_discovery"`
	HostIP        string             `mapstructure:"host_ip"`
	ServerURL     string             `mapstructure:"server_url"`
	DeviceManager map[string]Manager `mapstructure:"managers"`
	Hooks         []Hook             `mapstructure:"hooks"`
	Auth          Auth               `mapstructure:"auth"`
}
