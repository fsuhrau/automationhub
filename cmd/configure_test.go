package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"gopkg.in/yaml.v2"
)

// noManagers answers: autodetect ip, no manager.
const noManagers = "y\nn\nn\nn\nn\n"

func testEnv(t *testing.T, input, goos string) (configureEnv, *strings.Builder) {
	t.Helper()
	out := &strings.Builder{}
	return configureEnv{
		in:         strings.NewReader(input),
		out:        out,
		home:       t.TempDir(),
		goos:       goos,
		executable: func() (string, error) { return "/opt/bin/hub", nil },
	}, out
}

func readSaved(t *testing.T, env configureEnv, name string) config.Service {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(env.home, ".automationhub", name))
	if err != nil {
		t.Fatal(err)
	}
	var cfg config.Service
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestConfigureWritesTheDefaultConfig(t *testing.T) {
	env, out := testEnv(t, noManagers+"\n", "linux")

	if err := runConfigure(env); err != nil {
		t.Fatal(err)
	}

	cfg := readSaved(t, env, "config.yaml")
	wantDB := filepath.Join(env.home, ".automationhub", "sqlite.db")
	if !cfg.Autodetect || cfg.Database.SQLite == nil || cfg.Database.SQLite.Path != wantDB {
		t.Errorf("cfg = %+v", cfg)
	}
	if !strings.Contains(out.String(), "Setup finished!") || strings.Contains(out.String(), "LaunchAgent") {
		t.Errorf("out = %s", out)
	}
}

func TestConfigureWritesTheNamedConfig(t *testing.T) {
	env, out := testEnv(t, "n\n10.0.0.4\nn\nn\nn\nn\nmine.yaml\n", "linux")

	if err := runConfigure(env); err != nil {
		t.Fatal(err)
	}

	if cfg := readSaved(t, env, "mine.yaml"); cfg.HostIP != "10.0.0.4" {
		t.Errorf("cfg = %+v", cfg)
	}
	if !strings.Contains(out.String(), "hub master --config "+filepath.Join(env.home, ".automationhub", "mine.yaml")) {
		t.Errorf("out = %s", out)
	}
}

func TestConfigureOnMacInstallsALaunchAgent(t *testing.T) {
	env, out := testEnv(t, noManagers+"\ny\n", "darwin")

	if err := runConfigure(env); err != nil {
		t.Fatal(err)
	}

	plist := launchAgentPath(env.home)
	if _, err := os.Stat(plist); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "launchctl load "+plist) {
		t.Errorf("out = %s", out)
	}
}

func TestConfigureOnMacCanDeclineTheLaunchAgent(t *testing.T) {
	env, _ := testEnv(t, noManagers+"\nn\n", "darwin")

	if err := runConfigure(env); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(launchAgentPath(env.home)); !os.IsNotExist(err) {
		t.Errorf("plist exists: %v", err)
	}
}

func TestConfigureFailures(t *testing.T) {
	boom := errors.New("no executable")
	tests := map[string]struct {
		input string
		goos  string
		edit  func(*testing.T, *configureEnv)
		want  string
	}{
		"input ends early": {input: "y\n", goos: "linux", want: "read input"},
		"home is a file": {input: noManagers + "\n", goos: "linux", want: "create",
			edit: func(t *testing.T, e *configureEnv) {
				e.home = filepath.Join(t.TempDir(), "file")
				if err := os.WriteFile(e.home, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}},
		"config name is a directory": {input: noManagers + ".\n", goos: "linux", want: "write config"},
		"no executable": {input: noManagers + "\ny\n", goos: "darwin", want: "find executable",
			edit: func(_ *testing.T, e *configureEnv) { e.executable = func() (string, error) { return "", boom } }},
		"launch agent fails": {input: noManagers + "\ny\n", goos: "darwin", want: "create LaunchAgents directory",
			edit: func(t *testing.T, e *configureEnv) {
				if err := os.MkdirAll(filepath.Join(e.home, "Library"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(e.home, "Library", "LaunchAgents"), nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			env, _ := testEnv(t, tc.input, tc.goos)
			if tc.edit != nil {
				tc.edit(t, &env)
			}
			if err := runConfigure(env); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestConfigureCommandUsesTheCommandStreams(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	resetViper(t)
	root := newRootCmd()
	var out strings.Builder
	root.SetIn(strings.NewReader(noManagers + "\n"))
	root.SetOut(&out)
	root.SetArgs([]string{"--config", writeConfigFile(t, config.Service{}), "configure"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(home, ".automationhub", "config.yaml")); err != nil {
		t.Error(err)
	}
	if !strings.Contains(out.String(), "Welcome to the configuration wizard.") {
		t.Errorf("out = %s", &out)
	}
}
