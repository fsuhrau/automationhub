package hub

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/sirupsen/logrus"
)

// RunNode runs the device manager, keeps the connection to the master
// alive and serves http until the server fails.
func (s *Service) RunNode(ctx context.Context, handler *node.ReconnectHandler) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := s.deviceManager.Run(ctx, false); err != nil {
		return err
	}
	defer s.deviceManager.StopObserver()

	go handler.ObserveConnectionState(ctx, s.cfg, s.deviceManager.Managers())

	err := s.publicRouter.Run(fmt.Sprintf(":%d", s.cfg.Port))
	logrus.Infof("Stopping Server")
	return err
}
