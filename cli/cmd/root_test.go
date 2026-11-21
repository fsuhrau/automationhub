package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestVersionPrintsTheVersion(t *testing.T) {
	out, err := execute(t, "version")

	if err != nil || !strings.Contains(out, "dev") {
		t.Errorf("out = %q, err = %v", out, err)
	}
}

func TestRootHasTheTestAndVersionCommands(t *testing.T) {
	root := newRootCmd()
	for _, path := range [][]string{{"test"}, {"version"}, {"test", "run"}} {
		if c, _, err := root.Find(path); err != nil || c.Name() != path[len(path)-1] {
			t.Errorf("command %v: %v", path, err)
		}
	}
}

func TestInitConfigReportsTheConfigFileItUses(t *testing.T) {
	resetViper(t)
	path := emptyConfig(t)
	var stderr bytes.Buffer

	if err := initConfig(path, nil, &stderr); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stderr.String(), "Using config file: "+path) {
		t.Errorf("stderr = %q", &stderr)
	}
}

func TestInitConfigSearchesTheHomeDirectory(t *testing.T) {
	resetViper(t)
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".cli.yaml"), []byte("token: x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer

	if err := initConfig("", func() (string, error) { return home, nil }, &stderr); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stderr.String(), filepath.Join(home, ".cli.yaml")) || viper.GetString("token") != "x" {
		t.Errorf("stderr = %q", &stderr)
	}
}

func TestInitConfigIsSilentWithoutAConfigFile(t *testing.T) {
	resetViper(t)
	var stderr bytes.Buffer

	err := initConfig("", func() (string, error) { return t.TempDir(), nil }, &stderr)

	if err != nil || stderr.Len() != 0 {
		t.Errorf("err = %v, stderr = %q", err, &stderr)
	}
}

func TestInitConfigFailsWithoutHomeDirectory(t *testing.T) {
	resetViper(t)
	boom := errors.New("no home")

	err := initConfig("", func() (string, error) { return "", boom }, &bytes.Buffer{})

	if !errors.Is(err, boom) {
		t.Errorf("err = %v", err)
	}
}

func TestExitOnError(t *testing.T) {
	var codes []int
	var stderr bytes.Buffer
	exit := func(code int) { codes = append(codes, code) }

	exitOnError(nil, &stderr, exit)
	exitOnError(errors.New("boom"), &stderr, exit)

	if len(codes) != 1 || codes[0] != 1 || stderr.String() != "boom\n" {
		t.Errorf("codes = %v, stderr = %q", codes, &stderr)
	}
}

func TestExecuteRunsTheProcessArguments(t *testing.T) {
	resetViper(t)
	args := os.Args
	t.Cleanup(func() { os.Args = args })
	os.Args = []string{"cli", "--config", emptyConfig(t), "version"}

	Execute()
}
