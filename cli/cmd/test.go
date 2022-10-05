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
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/cli/api"
	"github.com/spf13/cobra"
	"strconv"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test http://localhost:8002 projectID appID",
	Long:  `list all available tests with id and name`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 4 {
			return fmt.Errorf("missing parameter")
		}
		var (
			apiURL    string
			projectID string
			appID     uint
		)
		apiURL = args[1]
		projectID = args[2]
		u, _ := strconv.ParseUint(args[3], 10, 64)
		appID = uint(u)

		client := api.NewClient(apiURL, apiToken, projectID, appID)

		tests, err := client.GetTests(context.Background())
		if err != nil {
			return err
		}

		for i := range tests {
			fmt.Printf("ID: %d Name: %s\n", tests[i].ID, tests[i].Name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
