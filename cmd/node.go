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
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device/node"
	"github.com/fsuhrau/automationhub/endpoints/manager"
	node_client "github.com/fsuhrau/automationhub/endpoints/node"
	"github.com/fsuhrau/automationhub/hub"
	node_hub "github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "start an automaton hub node and connect to the server",
	Long:  `an automation hub node will handle every connection on the local machine and will act as a proxy between local attached devices and the automation hub server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var serviceConfig config.Service

		if err := viper.Unmarshal(&serviceConfig); err != nil {
			return err
		}

		hostIP := getHostIP(serviceConfig)

		db, err := storage.GetDB(serviceConfig.Database)
		if err != nil {
			return err
		}

		logger := logrus.New()

		deviceStore := node.NewMemoryDeviceStore(serviceConfig.DeviceManager)
		deviceManager := hub.NewDeviceManager(logger, serviceConfig.MasterURL, serviceConfig.Identifier)

		reconnectHandler := node_hub.ReconnectHandler{}
		rpcNode := node_hub.NewRPCNode(serviceConfig, deviceManager, &reconnectHandler)

		server := hub.NewService(logger, hostIP, deviceManager, serviceConfig, deviceStore, db)

		server.AddEndpoint(manager.New(logger, deviceManager, serviceConfig))

		// endpoint for websocket connection
		server.AddEndpoint(deviceManager)

		var deviceManagers []string

		for k, _ := range serviceConfig.DeviceManager {
			deviceManagers = append(deviceManagers, k)
		}

		nodeClient := node_client.New(logger, db, rpcNode, deviceManager, serviceConfig, deviceManagers, &reconnectHandler)

		server.AddEndpoint(nodeClient)

		go nodeClient.ConnectAndServe()

		return server.RunNode(&reconnectHandler)
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
