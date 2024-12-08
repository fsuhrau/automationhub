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
	"github.com/fsuhrau/automationhub/endpoints/api"
	"github.com/fsuhrau/automationhub/endpoints/manager"
	node_master "github.com/fsuhrau/automationhub/endpoints/master"
	"github.com/fsuhrau/automationhub/endpoints/web"
	"github.com/fsuhrau/automationhub/hub"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
)

// masterCmd represents the run command
var masterCmd = &cobra.Command{
	Use:   "master",
	Short: "start the automaton hub master server",
	Long:  `automation hub server is a service which handle device connections and provides a device inspector gui`,
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

		deviceStore := storage.NewDeviceStore(db)

		logger := logrus.New()

		deviceManager := hub.NewDeviceManager(logger, "", "")
		sessionManager := hub.NewSessionManager(logger, deviceManager)
		nodeManager := hub.NewNodeManager(logger, db)

		server := hub.NewService(logger, hostIP, deviceManager, serviceConfig, deviceStore, db)

		server.AddEndpoint(api.New(logger, db, serviceConfig.NodeUrl, deviceManager, sessionManager, serviceConfig, nodeManager))
		server.AddEndpoint(manager.New(logger, deviceManager, serviceConfig))
		server.AddEndpoint(web.New(serviceConfig))
		server.AddEndpoint(node_master.New(serviceConfig, deviceManager, nodeManager, nil))

		// endpoint for websocket connection
		server.AddEndpoint(deviceManager)

		server.RegisterHooks(serviceConfig.Hooks)

		return server.RunMaster(nodeManager, sessionManager)
	},
}

func getHostIP(cfg config.Service) net.IP {
	var hostIP net.IP
	if len(cfg.HostIP) > 0 {
		hostIP = net.ParseIP(cfg.HostIP)
	}

	if hostIP == nil {
		hostIP = utils.GetOutboundIP()
	}
	return hostIP
}

func init() {
	rootCmd.AddCommand(masterCmd)
}
