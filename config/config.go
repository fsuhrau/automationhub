package config

type WebDriver struct {
	BundleID string `yaml:"bundleId,omitempty" mapstructure:"bundleId"`
}

type Device struct {
	Identifier string            `yaml:"identifier,omitempty" mapstructure:"identifier"`
	Parameter  map[string]string `yaml:"parameter,omitempty" mapstructure:"parameter"`
}

type Manager struct {
	Enabled          bool              `yaml:"enabled,omitempty" mapstructure:"enabled"`
	UseOSScreenshot  bool              `yaml:"use_os_screenshot,omitempty" mapstructure:"use_os_screenshot"`
	WebDriver        *WebDriver        `yaml:"webdriver,omitempty" mapstructure:"webdriver"`
	UnityPath        string            `yaml:"unity_path,omitempty" mapstructure:"unity_path"`
	UnityBuildTarget string            `yaml:"unity_build_target,omitempty" mapstructure:"unity_build_target"`
	Devices          []Device          `yaml:"devices,omitempty" mapstructure:"devices"`
	Browser          map[string]string `yaml:"browser,omitempty" mapstructure:"browser"`
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

type Password struct {
	Secret string `yaml:"secret,omitempty" mapstructure:"secret"`
}

type Auth struct {
	Token    *Token    `yaml:"token,omitempty" mapstructure:"token"`
	Password *Password `yaml:"token,omitempty" mapstructure:"password"`
	OAuth2   *OAuth2   `yaml:"oauth2,omitempty" mapstructure:"oauth2"`
	Github   *Github   `yaml:"github,omitempty" mapstrucutre:"github"`
}

func (a *Auth) AuthenticationRequired() bool {
	return a.OAuth2 != nil || a.Github != nil || a.Token != nil || a.Password != nil
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
	Identifier    string             `yaml:"identifier,omitempty" mapstructure:"identifier"`
	Port          int32              `yaml:"port,omitempty" mapstructure:"port"`
	Autodetect    bool               `yaml:"autodetect_ip,omitempty" mapstructure:"autodetect_ip"`
	AutoDiscovery bool               `yaml:"auto_discovery,omitempty" mapstructure:"auto_discovery"`
	HostIP        string             `yaml:"host_ip,omitempty" mapstructure:"host_ip"`
	NodeUrl       string             `yaml:"node_url,omitempty" mapstructure:"node_url"`
	MasterURL     string             `yaml:"master_url,omitempty" mapstructure:"master_url"`
	Cors          []string           `yaml:"cors,omitempty" mapstructure:"cors"`
	DeviceManager map[string]Manager `yaml:"managers,omitempty" mapstructure:"managers"`
	Hooks         []Hook             `yaml:"hooks,omitempty" mapstructure:"hooks"`
	Auth          Auth               `yaml:"auth,omitempty" mapstructure:"auth"`
	Database      Database           `yaml:"database,omitempty" mapstructure:"database"`
}
