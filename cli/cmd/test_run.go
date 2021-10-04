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
)

var (
	appPath string
	appID   int
	testID  int
	params  string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a new test",
	Long: `Run a new test.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {

		client := api.NewClient(apiURL, apiKey)
		if len(appPath) > 0 {
			app, err := client.UploadApp(appPath)
			if err != nil {
				return err
			}
			appID = int(app.ID)
		}

		testRun, err := client.ExecuteTest(context.Background(), testID, appID, params)
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("find your results at: http://%s/test/%d/run/%d", apiURL, testID, testRun.ID))
		return nil
	},
}

func init() {
	testCmd.AddCommand(runCmd)

	testCmd.PersistentFlags().StringVar(&appPath, "app", "", "app path")
	testCmd.PersistentFlags().IntVar(&appID, "appid", 0, "appid")
	testCmd.PersistentFlags().IntVar(&testID, "testid", 0, "testid")
	testCmd.PersistentFlags().StringVar(&params, "params", "", "test environment parameter")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
