package hub

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/sirupsen/logrus"
)

func (s *Service) RunNode(handler *node.ReconnectHandler) error {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if len(s.cfg.MasterURL) == 0 {
		return fmt.Errorf("slave need a master url to connect to")
	}

	s.RegisterDeviceHandler(s.cfg.MasterURL)

	if err := s.deviceManager.Run(ctx, false); err != nil {
		return err
	}
	defer s.deviceManager.StopObserver()

	go handler.ObserveConnectionState(ctx, s.cfg, s.deviceManager.Managers())

	err := s.router.Run(fmt.Sprintf(":%d", s.cfg.Port))
	logrus.Infof("Stopping Server")
	if err != nil {
		return err
	}
	return nil
}
