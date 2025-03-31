package hub

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/device/node"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (s *Service) RunMaster(nodeManager *NodeManager, sessionManager *SessionManager) error {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if d, ok := s.cfg.DeviceManager[node.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager for remote nodes")
		s.deviceManager.AddHandler(node.NewHandler(s.logger, s.sd, nodeManager, s.db))
	}

	if err := s.deviceManager.Run(ctx, true); err != nil {
		return err
	}
	defer s.deviceManager.StopObserver()

	sessionManager.Run(ctx)

	if nodeManager != nil {
		nodeManager.Run(ctx)
	}

	s.publicRouter.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	if s.cfg.Port == 0 {
		s.cfg.Port = 8002
	}

	runOn := fmt.Sprintf(":%d", s.cfg.Port)
	err := s.publicRouter.Run(runOn)
	logrus.Infof("Stopping Server")
	if err != nil {
		return err
	}

	return nil
}
