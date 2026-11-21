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
	"github.com/spf13/cobra"
)

// newTestCmd lists the tests of an app; its run subcommand starts one.
func newTestCmd(global *globalOptions) *cobra.Command {
	opts := &runOptions{}
	cmd := &cobra.Command{
		Use:   "test",
		Short: "test http://localhost:8002 projectID appID",
		Long:  `list all available tests with id and name`,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := parseTarget(args)
			if err != nil {
				return err
			}
			tests, err := t.client(global.apiToken).GetTests(cmd.Context())
			if err != nil {
				return err
			}
			for i := range tests {
				fprintf(cmd.OutOrStdout(), "ID: %d Name: %s\n", tests[i].ID, tests[i].Name)
			}
			return nil
		},
	}
	cmd.PersistentFlags().BoolVarP(&opts.async, "async", "a", false, "run command async observe status manually")
	cmd.AddCommand(newRunCmd(global, opts))
	return cmd
}
