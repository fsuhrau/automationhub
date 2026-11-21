package cmd

import (
	"errors"
	"fmt"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device/node"
	"github.com/fsuhrau/automationhub/endpoints/manager"
	node_client "github.com/fsuhrau/automationhub/endpoints/node"
	"github.com/fsuhrau/automationhub/hub"
	node_hub "github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// nodeService is the wired node, ready to run.
type nodeService struct {
	server  *hub.Service
	handler *node_hub.ReconnectHandler
}

func newNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "node",
		Short: "start an automaton hub node and connect to the server",
		Long:  `an automation hub node will handle every connection on the local machine and will act as a proxy between local attached devices and the automation hub server`,
		RunE: func(*cobra.Command, []string) error {
			cfg, err := loadServiceConfig()
			if err != nil {
				return err
			}
			return runNode(cfg)
		},
	}
}

func runNode(cfg config.Service) error {
	n, err := buildNode(cfg)
	if err != nil {
		return err
	}
	return errors.Join(n.handler.HandleConnect(), n.server.RunNode(n.handler))
}

// buildNode wires the node; the service is only usable when err is nil.
func buildNode(cfg config.Service) (*nodeService, error) {
	db, err := storage.GetDB(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	logger := logrus.New()
	deviceManager := hub.NewDeviceManager(logger, cfg)
	handler := &node_hub.ReconnectHandler{}
	rpcNode := node_hub.NewRPCNode(cfg, deviceManager, handler)
	server := hub.NewService(logger, getHostIP(cfg), deviceManager, cfg, node.NewMemoryDeviceStore(cfg.DeviceManager), db)

	nodeClient := node_client.New(logger, db, rpcNode, deviceManager, cfg, managerNames(cfg), handler)
	handler.ConnectFunction = nodeClient.ConnectAndServe
	return &nodeService{server: server, handler: handler}, addEndpoints(server, manager.New(logger, deviceManager, cfg), deviceManager, nodeClient)
}

func managerNames(cfg config.Service) []string {
	names := make([]string, 0, len(cfg.DeviceManager))
	for name := range cfg.DeviceManager {
		names = append(names, name)
	}
	return names
}
