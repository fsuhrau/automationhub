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
	"github.com/fsuhrau/automationhub/endpoints/selenium"
	"github.com/fsuhrau/automationhub/endpoints/web"
	"github.com/fsuhrau/automationhub/hub"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/utils"
	"github.com/getsentry/sentry-go"
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
		sentryDns := viper.GetString("sentry_dsn")
		if len(sentryDns) > 0 {
			sentry.Init(sentry.ClientOptions{
				Dsn: sentryDns,
			})
		}

		var serviceConfig config.Service
		if err := viper.Unmarshal(&serviceConfig); err != nil {
			return err
		}

		hostIP := getHostIP(serviceConfig)

		db, err := storage.GetDB()
		if err != nil {
			return err
		}

		ds := storage.NewDeviceStore(db)

		logger := logrus.New()
		deviceManager := hub.NewDeviceManager(logger, db)
		sessionManager := hub.NewSessionManager(logger, deviceManager)
		server := hub.NewService(logger, hostIP, deviceManager, sessionManager, serviceConfig, ds)
		server.AddEndpoint(web.New(serviceConfig))
		server.AddEndpoint(api.New(logger, db, hostIP, deviceManager, sessionManager, serviceConfig))
		server.AddEndpoint(selenium.New(logger, nil, deviceManager, sessionManager))
		// server.AddEndpoint(inspector.New(logger, deviceManager, sessionManager))
		server.RegisterHooks(serviceConfig.Hooks)

		return server.RunMaster()
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

	if false {
		masterCmd.Flags().StringP("address", "a", "", "address to listen on")
	}

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// masterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
