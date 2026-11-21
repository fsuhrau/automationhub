package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fsuhrau/automationhub/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const defaultConfigName = "config.yaml"

// configureEnv is everything the configure wizard takes from its environment.
type configureEnv struct {
	in         io.Reader
	out        io.Writer
	home       string
	goos       string
	executable func() (string, error)
}

func newConfigureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "create a configuration to run the hub",
		Long:  `wizard to create a configuration to your needs.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runConfigure(configureEnv{
				in:         cmd.InOrStdin(),
				out:        cmd.OutOrStdout(),
				home:       os.Getenv("HOME"),
				goos:       runtime.GOOS,
				executable: os.Executable,
			})
		},
	}
}

func runConfigure(env configureEnv) error {
	w := newWizard(env.in, env.out)
	cfg, err := w.collect()
	if err != nil {
		return err
	}
	dir := filepath.Join(env.home, ".automationhub")
	cfg.Database = config.Database{SQLite: &config.SQLite{Path: filepath.Join(dir, "sqlite.db")}}

	w.say(rule, "Save your config!", "")
	name := w.readLine("Enter a name for the config: (default: " + defaultConfigName + ")")
	if name == "" {
		name = defaultConfigName
	}
	configPath, err := saveConfig(dir, name, cfg)
	if err != nil {
		return err
	}
	w.say("", "Setup finished!", "your config was written to "+configPath, "",
		fmt.Sprintf("you can start the hub now via 'hub master --config %s'", configPath))
	if env.goos == "darwin" {
		w.offerLaunchAgent(env, configPath)
	}
	return w.err
}

// saveConfig writes cfg as YAML to dir/name and returns that path.
func saveConfig(dir, name string, cfg config.Service) (string, error) {
	data, _ := yaml.Marshal(&cfg) // a struct of strings, bools and maps always marshals
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}
	return path, nil
}

func (w *wizard) offerLaunchAgent(env configureEnv, configPath string) {
	w.say(rule)
	if !w.askYes("Would you like to create a LaunchAgent?") {
		return
	}
	exe, err := env.executable()
	if err != nil {
		w.err = fmt.Errorf("find executable: %w", err)
		return
	}
	plist := launchAgentPath(env.home)
	if err := writeLaunchAgent(plist, newLaunchAgent(filepath.Dir(exe), configPath)); err != nil {
		w.err = err
		return
	}
	w.say("Agent installed you can start it via:", "launchctl load "+plist)
}
