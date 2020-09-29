// Copyright © 2020 Fabian Suhrau <fabian.suhrau@me.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/fsuhrau/automationhub/hub"
	"github.com/spf13/cobra"
)

// slaveCmd represents the slave command
var slaveCmd = &cobra.Command{
	Use:   "slave",
	Short: "start an automaton hub slave and connect to the server",
	Long:  `an automation hub slave will handle every connection on the local machine and will act as a proxy between local attached devices and the automation hub server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		server := hub.NewService()
		return server.RunSlave()
	},
}

func init() {
	rootCmd.AddCommand(slaveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// slaveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// slaveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
