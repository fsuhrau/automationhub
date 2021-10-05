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
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/cli/api"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/utils/sync"
	"github.com/r3labs/sse/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

var (
	appPath string
	appID   int
	testID  int
	params  string
	async *bool
	success bool
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run --url http://localhost:8002 --appid 50 --testid 9 --async",
	Long: `Run a new test.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(apiURL, apiKey)
		if len(appPath) > 0 {
			logrus.Info("uploading app")
			app, err := client.UploadApp(appPath)
			if err != nil {
				return err
			}
			appID = int(app.ID)
			logrus.Infof("upload finished new appId: %d", appID)
		}

		logrus.Infof("execute test %d with appId: %d", testID, appID)
		testRun, err := client.ExecuteTest(context.Background(), testID, appID, params)
		if err != nil {
			return err
		}
		defer func() {
			logrus.Infof("Check your restults at: %s/web/test/%d/run/%d", apiURL, testID, testRun.ID)
		}()
		if async != nil && *async == true {
			logrus.Infof("Test Running...")
		} else {
			eventsChannel := make(chan *sse.Event)
			client := sse.NewClient(fmt.Sprintf("%s/api/sse", apiURL))
			client.SubscribeChan(fmt.Sprintf("test_run_%d_finished", testRun.ID), eventsChannel)
			wg := sync.ExtendedWaitGroup{}
			wg.Add(1)
			go waitForResult(&wg, eventsChannel)
			close(eventsChannel)
			if err := wg.WaitWithTimeout(5 * time.Minute); err != nil {
				return err
			}
			if success {
				logrus.Infof("Test finished Successfully!")
			}
		}
		return nil
	},
}

func waitForResult(wg *sync.ExtendedWaitGroup, eventsChannel chan *sse.Event) {
	for {
		select {
		case event := <- eventsChannel:
			var finishedEvent events.TestRunFinishedPayload
			json.Unmarshal(event.Data, &finishedEvent)
			success = finishedEvent.Success
			wg.Done()
			return
		}
	}
}

func init() {
	success = false
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
	async = testCmd.PersistentFlags().BoolP("async", "a", false, "run command async observe status manually")
}
