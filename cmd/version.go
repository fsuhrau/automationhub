package cmd

import (
	"fmt"

	"github.com/fsuhrau/automationhub/hub"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print current app version",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), hub.GetVersion())
		},
	}
}
