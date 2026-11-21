package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const authYAML = `
auth:
  password:
    secret: pw-secret
  github:
    redirect_url: http://gh/cb
    credentials: gh.json
    scopes: [user, repo]
    secret: gh-secret
`

func TestAuthUnmarshalViper(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(authYAML)); err != nil {
		t.Fatal(err)
	}
	var cfg Service
	if err := v.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	assertAuth(t, cfg.Auth)
}

func TestAuthUnmarshalYAML(t *testing.T) {
	var cfg Service
	if err := yaml.Unmarshal([]byte(authYAML), &cfg); err != nil {
		t.Fatal(err)
	}
	assertAuth(t, cfg.Auth)
}

func assertAuth(t *testing.T, a Auth) {
	t.Helper()
	if a.Password == nil || a.Password.Secret != "pw-secret" {
		t.Errorf("password not filled: %+v", a.Password)
	}
	want := &Github{RedirectUrl: "http://gh/cb", Credentials: "gh.json", Scopes: []string{"user", "repo"}, Secret: "gh-secret"}
	if !reflect.DeepEqual(a.Github, want) {
		t.Errorf("github = %+v, want %+v", a.Github, want)
	}
}

func TestAuthenticationRequired(t *testing.T) {
	tests := map[string]struct {
		auth Auth
		want bool
	}{
		"none":     {Auth{}, false},
		"oauth2":   {Auth{OAuth2: &OAuth2{}}, true},
		"github":   {Auth{Github: &Github{}}, true},
		"password": {Auth{Password: &Password{}}, true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tc.auth.AuthenticationRequired(); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

const fullYAML = `
log: debug
identifier: hub-1
port: 8002
autodetect_ip: true
auto_discovery: true
host_ip: 10.0.0.1
node_url: http://node
master_url: http://master
cors: [http://a, http://b]
managers:
  android:
    enabled: true
    use_os_screenshot: true
    unity_path: /unity
    unity_build_target: Android
    devices:
      - identifier: dev1
        parameter: {k: v}
    browser: {chrome: /bin/chrome}
  ios:
    enabled: true
    webdriver:
      bundleId: com.example
      launcher: go-ios
      xctestrun: a.xctestrun
      project: wda.xcodeproj
      scheme: Runner
      derived_data: /dd
hooks:
  - provider: slack
    url: http://slack
    token: tok
    username: bot
    channel: "#ci"
auth:
  token: {auth_token: tk}
  oauth2:
    auth_url: http://auth
    token_url: http://token
    redirect_url: http://redirect
    user_url: http://user
    credentials: c.json
    scopes: [a]
    secret: s
database:
  sqlite: {path: /tmp/db}
  postgres: {host: h, port: "5432", user: u, password: p, database: d}
`

func fullConfig() Service {
	return Service{
		LogLevel: "debug", Identifier: "hub-1", Port: 8002, Autodetect: true, AutoDiscovery: true,
		HostIP: "10.0.0.1", NodeUrl: "http://node", MasterURL: "http://master", Cors: []string{"http://a", "http://b"},
		DeviceManager: map[string]Manager{
			"android": {
				Enabled: true, UseOSScreenshot: true, UnityPath: "/unity", UnityBuildTarget: "Android",
				Devices: []Device{{Identifier: "dev1", Parameter: map[string]string{"k": "v"}}},
				Browser: map[string]string{"chrome": "/bin/chrome"},
			},
			"ios": {Enabled: true, WebDriver: &WebDriver{
				BundleID: "com.example", Launcher: "go-ios", XCTestRun: "a.xctestrun",
				Project: "wda.xcodeproj", Scheme: "Runner", DerivedData: "/dd",
			}},
		},
		Hooks: []Hook{{Provider: "slack", Url: "http://slack", Token: "tok", Username: "bot", Channel: "#ci"}},
		Auth: Auth{
			Token: &Token{AuthToken: "tk"},
			OAuth2: &OAuth2{
				AuthUrl: "http://auth", TokenUrl: "http://token", RedirectUrl: "http://redirect",
				UserUrl: "http://user", Credentials: "c.json", Scopes: []string{"a"}, Secret: "s",
			},
		},
		Database: Database{
			SQLite:   &SQLite{Path: "/tmp/db"},
			Postgres: &Postgres{Host: "h", Port: "5432", User: "u", Password: "p", Database: "d"},
		},
	}
}

func TestEverySectionUnmarshalsFromViper(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(fullYAML)); err != nil {
		t.Fatal(err)
	}
	var got Service
	if err := v.Unmarshal(&got); err != nil {
		t.Fatal(err)
	}
	if want := fullConfig(); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
}

func TestEverySectionRoundTripsThroughYAML(t *testing.T) {
	data, err := yaml.Marshal(fullConfig())
	if err != nil {
		t.Fatal(err)
	}
	var got Service
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if want := fullConfig(); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
}

func TestEmptyConfigHasZeroDefaults(t *testing.T) {
	var cfg Service
	if err := yaml.Unmarshal([]byte("{}"), &cfg); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg, Service{}) || cfg.Auth.AuthenticationRequired() {
		t.Errorf("not the zero value: %+v", cfg)
	}
	out, err := yaml.Marshal(Service{})
	if err != nil || strings.TrimSpace(string(out)) != "{}" {
		t.Errorf("empty config marshals to %q, %v", out, err)
	}
}
