package cmd

import (
	"fmt"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/endpoints/api"
	"github.com/fsuhrau/automationhub/endpoints/manager"
	node_master "github.com/fsuhrau/automationhub/endpoints/master"
	"github.com/fsuhrau/automationhub/endpoints/web"
	"github.com/fsuhrau/automationhub/hub"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// masterService is the wired master, ready to run.
type masterService struct {
	server   *hub.Service
	nodes    *hub.NodeManager
	sessions *hub.SessionManager
}

func newMasterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "master",
		Short: "start the automaton hub master server",
		Long:  `automation hub server is a service which handle device connections and provides a device inspector gui`,
		RunE: func(*cobra.Command, []string) error {
			cfg, err := loadServiceConfig()
			if err != nil {
				return err
			}
			return runMaster(cfg)
		},
	}
}

func runMaster(cfg config.Service) error {
	m, err := buildMaster(cfg)
	if err != nil {
		return err
	}
	return m.server.RunMaster(m.nodes, m.sessions)
}

// buildMaster wires the master; the service is only usable when err is nil.
func buildMaster(cfg config.Service) (*masterService, error) {
	db, err := storage.GetDB(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	logger := logrus.New()
	deviceManager := hub.NewDeviceManager(logger, cfg)
	m := &masterService{
		sessions: hub.NewSessionManager(logger, deviceManager),
		nodes:    hub.NewNodeManager(logger, db),
		server:   hub.NewService(logger, getHostIP(cfg), deviceManager, cfg, storage.NewDeviceStore(db), db),
	}
	m.server.RegisterHooks(cfg.Hooks)
	return m, addEndpoints(m.server,
		api.New(logger, db, cfg.NodeUrl, deviceManager, m.sessions, cfg, m.nodes),
		manager.New(logger, deviceManager, cfg),
		web.New(db, cfg),
		node_master.New(cfg, deviceManager, m.nodes, nil),
		// endpoint for websocket connection
		deviceManager,
	)
}
