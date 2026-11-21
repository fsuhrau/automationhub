package cmd

import (
	"context"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/wiring"
	"github.com/spf13/cobra"
)

func newMasterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "master",
		Short: "start the automaton hub master server",
		Long:  `automation hub server is a service which handle device connections and provides a device inspector gui`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadServiceConfig()
			if err != nil {
				return err
			}
			return runMaster(cmd.Context(), cfg)
		},
	}
}

// runMaster builds the master of the config and runs it until the command
// is cancelled or the server fails.
func runMaster(ctx context.Context, cfg config.Service) error {
	m, err := wiring.BuildMaster(cfg)
	if err != nil {
		return err
	}
	return m.Run(ctx)
}
