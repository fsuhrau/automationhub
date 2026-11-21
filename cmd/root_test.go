package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/config"
	"github.com/spf13/viper"
)

func TestVersionPrintsTheVersion(t *testing.T) {
	resetViper(t)
	out, err := execute(t, "--config", writeConfigFile(t, config.Service{}), "version")
	if err != nil || strings.TrimSpace(out) != "dev" {
		t.Errorf("out = %q, err = %v", out, err)
	}
}

func TestRootHasAllSubcommands(t *testing.T) {
	root := newRootCmd()
	for _, name := range []string{"configure", "master", "node", "version", "mcp-proxy"} {
		if c, _, err := root.Find([]string{name}); err != nil || c.Name() != name {
			t.Errorf("command %s: %v", name, err)
		}
	}
}

func TestConfigFlagFeedsViper(t *testing.T) {
	resetViper(t)
	path := writeConfigFile(t, config.Service{Identifier: "from-file"})

	if _, err := execute(t, "--config", path, "version"); err != nil {
		t.Fatal(err)
	}

	if got := viper.GetString("identifier"); got != "from-file" {
		t.Errorf("identifier = %q", got)
	}
}

func TestInitConfigSearchesTheHomeDirectory(t *testing.T) {
	resetViper(t)
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".automationhub.yaml"), []byte("identifier: from-home\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := initConfig("", func() (string, error) { return home, nil }); err != nil {
		t.Fatal(err)
	}

	if got := viper.GetString("identifier"); got != "from-home" {
		t.Errorf("identifier = %q", got)
	}
}

func TestInitConfigFailsWithoutHomeDirectory(t *testing.T) {
	resetViper(t)
	boom := errors.New("no home")

	err := initConfig("", func() (string, error) { return "", boom })

	if !errors.Is(err, boom) {
		t.Errorf("err = %v", err)
	}
}

func TestExitOnError(t *testing.T) {
	var codes []int
	exit := func(code int) { codes = append(codes, code) }

	exitOnError(nil, exit)
	exitOnError(errors.New("boom"), exit)

	if len(codes) != 1 || codes[0] != 1 {
		t.Errorf("exit codes = %v", codes)
	}
}

func TestExecuteRunsTheProcessArguments(t *testing.T) {
	resetViper(t)
	args := os.Args
	t.Cleanup(func() { os.Args = args })
	os.Args = []string{"hub", "--config", writeConfigFile(t, config.Service{}), "version"}

	Execute()
}
