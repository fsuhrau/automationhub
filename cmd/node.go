package cmd

import (
	"context"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/wiring"
	"github.com/spf13/cobra"
)

func newNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "node",
		Short: "start an automaton hub node and connect to the server",
		Long:  `an automation hub node will handle every connection on the local machine and will act as a proxy between local attached devices and the automation hub server`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadServiceConfig()
			if err != nil {
				return err
			}
			return runNode(cmd.Context(), cfg)
		},
	}
}

// runNode builds the node of the config and runs it until the command is
// cancelled or the server fails.
func runNode(ctx context.Context, cfg config.Service) error {
	n, err := wiring.BuildNode(cfg)
	if err != nil {
		return err
	}
	return n.Run(ctx)
}
