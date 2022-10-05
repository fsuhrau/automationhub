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
	sse "github.com/r3labs/sse/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
	"time"
)

var (
	appPath  string
	binaryID int
	params   string
	async    *bool
	success  bool
	tags     string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run http://localhost:8002 projectID appID testName --binaryID 50 --binary path_to_app --tags \"tag1,tag2,tag3\" --params \"param1=1;parameter2=2\" --async",
	Long: `Run a new test.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 4 {
			return fmt.Errorf("missing parameter")
		}
		var (
			testID    uint
			apiURL    string
			projectID string
			appID     uint
			testName  string
		)

		apiURL = args[0]
		projectID = args[1]
		u, _ := strconv.ParseUint(args[2], 10, 64)
		appID = uint(u)
		testName = args[3]

		client := api.NewClient(apiURL, apiToken, projectID, appID)

		if len(appPath) > 0 {
			logrus.Info("uploading appBundle")
			appBundle, err := client.UploadBundle(appPath)
			if err != nil {
				return err
			}
			binaryID = int(appBundle.ID)
			logrus.Infof("upload finished new binaryID: %d", binaryID)

			if len(tags) > 0 {
				appBundle.Tags = tags
				if err := client.UpdateApp(context.Background(), appBundle); err != nil {
					return err
				}
			}
		}

		if len(testName) > 0 {
			tests, err := client.GetTests(context.Background())
			if err != nil {
				return err
			}
			for _, t := range tests {
				if t.Name == testName {
					testID = t.ID
				}
			}
		}

		if testID == 0 {
			return fmt.Errorf("no test provided or test could not be found")
		}

		logrus.Infof("execute test %d with binaryId: %d\n%s", testID, binaryID, params)
		parameter := strings.Split(params, ";")
		logrus.Infof("execute test %d with binaryId: %d", testID, binaryID)
		testRun, err := client.ExecuteTest(context.Background(), testID, binaryID, strings.Join(parameter, "\n"))
		if err != nil {
			return err
		}
		defer func() {
			logrus.Infof("Check your restults at: %s/project/%s/app/%d/test/%d/run/%d", apiURL, projectID, appID, testID, testRun.ID)
		}()
		if async != nil && *async == true {
			logrus.Infof("Test Running...")
		} else {
			client := sse.NewClient(fmt.Sprintf("%s/api/sse", apiURL))

			eventsChannel := make(chan *sse.Event)
			if err := client.SubscribeChan(fmt.Sprintf("test_run_%d_finished", testRun.ID), eventsChannel); err != nil {
				return err
			}

			runLogChannel := make(chan *sse.Event)
			if err := client.SubscribeChan(fmt.Sprintf("test_run_%d_log", testRun.ID), runLogChannel); err != nil {
				return err
			}

			wg := sync.ExtendedWaitGroup{}
			wg.Add(1)
			go waitForResult(&wg, eventsChannel, runLogChannel)
			if err := wg.WaitWithTimeout(5 * time.Minute); err != nil {
				return err
			}
			close(eventsChannel)
			if success {
				logrus.Infof("Test finished Successfully!")
			}
		}
		return nil
	},
}

func waitForResult(wg *sync.ExtendedWaitGroup, eventsChannel, runLogChannel chan *sse.Event) {
	for {
		select {
		case event := <-eventsChannel:
			if event != nil {
				var finishedEvent events.TestRunFinishedPayload
				json.Unmarshal(event.Data, &finishedEvent)
				success = finishedEvent.Success
				wg.Done()
				return
			}
		case log := <-runLogChannel:
			var ev events.NewTestLogEntryPayload
			json.Unmarshal(log.Data, &ev)
			logrus.Infof("log: %v", ev)
		}
	}
}

func init() {
	success = false
	testCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().StringVar(&appPath, "binary", "", "binary /path/to/apk/ipa/zip")
	runCmd.PersistentFlags().IntVar(&binaryID, "binaryID", 0, "binaryID 123")
	runCmd.PersistentFlags().StringVar(&params, "params", "", "params \"param1=1;param2=2\"")
	runCmd.PersistentFlags().StringVar(&tags, "tags", "", "tag \"tag1,tag2,tag3\"")
	async = testCmd.PersistentFlags().BoolP("async", "a", false, "run command async observe status manually")
}
