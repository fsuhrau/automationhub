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
