package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newRootCmd builds the command tree.
func newRootCmd() *cobra.Command {
	var cfgFile string
	root := &cobra.Command{
		Use: "hub",
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return initConfig(cfgFile, homedir.Dir)
		},
	}
	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.automationhub.yaml)")
	root.AddCommand(newConfigureCmd(), newMasterCmd(), newNodeCmd(), newVersionCmd(), newMCPProxyCmd())
	return root
}

// Execute runs the command line of the process. It is called by main.main().
func Execute() {
	exitOnError(newRootCmd().Execute(), os.Exit)
}

func exitOnError(err error, exit func(int)) {
	if err != nil {
		fmt.Println(err)
		exit(1)
	}
}

// initConfig reads the config file and the environment variables into viper.
func initConfig(cfgFile string, homeDir func() (string, error)) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else if err := useHomeConfig(homeDir); err != nil {
		return err
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()
	return nil
}

// useHomeConfig searches the config ".automationhub" in the home directory.
func useHomeConfig(homeDir func() (string, error)) error {
	home, err := homeDir()
	if err != nil {
		return fmt.Errorf("find home directory: %w", err)
	}
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".automationhub")
	return nil
}
