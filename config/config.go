package config

type WebDriver struct {
	BundleID string `yaml:"bundleId,omitempty" mapstructure:"bundleId"`
}

type Manager struct {
	Enabled         bool       `yaml:"enabled,omitempty" mapstructure:"enabled"`
	UseOSScreenshot bool       `yaml:"use_os_screenshot,omitempty" mapstructure:"use_os_screenshot"`
	WebDriver       *WebDriver `yaml:"webdriver,omitempty" mapstructure:"webdriver"`
}

type Hook struct {
	Provider string `yaml:"provider,omitempty" mapstructure:"provider"`
	Url      string `yaml:"url,omitempty" mapstructure:"url"`
	Token    string `yaml:"token,omitempty" mapstructure:"token"`
	Username string `yaml:"username,omitempty" mapstructure:"username"`
	Channel  string `yaml:"channel,omitempty" mapstructure:"channel"`
}

type Token struct {
}

type Auth struct {
	Token  *Token  `yaml:"token,omitempty" mapstructure:"token"`
	OAuth2 *OAuth2 `yaml:"oauth2,omitempty" mapstructure:"oauth2"`
	Github *Github `yaml:"github,omitempty" mapstrucutre:"github"`
}

func (a *Auth) AuthenticationRequired() bool {
	return a.OAuth2 != nil || a.Github != nil || a.Token != nil
}

type OAuth2 struct {
	AuthUrl     string   `yaml:"auth_url,omitempty" mapstructure:"auth_url"`
	TokenUrl    string   `yaml:"token_url,omitempty" mapstructure:"token_url"`
	RedirectUrl string   `yaml:"redirect_url,omitempty" mapstructure:"redirect_url"`
	UserUrl     string   `yaml:"user_url,omitempty" mapstructure:"user_url"`
	Credentials string   `yaml:"credentials,omitempty" mapstructure:"credentials"`
	Scopes      []string `yaml:"scopes,omitempty" mapstructure:"scopes"`
	Secret      string   `yaml:"secret,omitempty" mapstructure:"secret"`
}

type Github struct {
	RedirectUrl string   `yaml:"redirect_url,omitempty" mapstructure:"redirect_url"`
	Credentials string   `yaml:"credentials,omitempty" mapstructure:"credentials"`
	Scopes      []string `yaml:"scopes,omitempty" mapstructure:"scopes"`
	Secret      string   `yaml:"secret,omitempty" mapstructure:"secret"`
}

type Database struct {
	SQLiteDBPath string `yaml:"sqlite_db_path,omitempty" mapstructure:"sqlite_db_path"`
}

type Service struct {
	Port          uint16             `yaml:"port,omitempty" mapstructure:"port"`
	Autodetect    bool               `yaml:"autodetect_ip,omitempty" mapstructure:"autodetect_ip"`
	AutoDiscovery bool               `yaml:"auto_discovery,omitempty" mapstructure:"auto_discovery"`
	HostIP        string             `yaml:"host_ip,omitempty" mapstructure:"host_ip"`
	ServerURL     string             `yaml:"server_url,omitempty" mapstructure:"server_url"`
	DeviceManager map[string]Manager `yaml:"managers,omitempty" mapstructure:"managers"`
	Hooks         []Hook             `yaml:"hooks,omitempty" mapstructure:"hooks"`
	Auth          Auth               `yaml:"auth,omitempty" mapstructure:"auth"`
	Database      Database           `yaml:"database,omitempty" mapstructure:"database"`
}
