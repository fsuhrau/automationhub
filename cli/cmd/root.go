/*
Copyright © 2021 Fabian Suhrau <fabian.suhrau@me.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// globalOptions are the flags every command shares.
type globalOptions struct {
	apiToken string
}

// newRootCmd builds the command tree.
func newRootCmd() *cobra.Command {
	var (
		cfgFile string
		global  globalOptions
	)
	root := &cobra.Command{
		Use: "cli",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return initConfig(cfgFile, homedir.Dir, cmd.ErrOrStderr())
		},
	}
	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")
	root.PersistentFlags().StringVar(&global.apiToken, "token", "", "token for authentication")
	root.AddCommand(newTestCmd(&global), newVersionCmd())
	return root
}

// Execute runs the command line of the process. It is called by main.main().
func Execute() {
	exitOnError(newRootCmd().Execute(), os.Stderr, os.Exit)
}

// fprintf writes to out; a failing terminal is not worth reporting.
func fprintf(out io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(out, format, args...)
}

func exitOnError(err error, stderr io.Writer, exit func(int)) {
	if err != nil {
		fprintf(stderr, "%v\n", err)
		exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cfgFile string, homeDir func() (string, error), stderr io.Writer) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homeDir()
		if err != nil {
			return fmt.Errorf("find home directory: %w", err)
		}
		// Search config in home directory with name ".cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fprintf(stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
	return nil
}
