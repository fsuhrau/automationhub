package hub

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// defaultMasterPort is the port a master listens on without a configured one.
const defaultMasterPort = 8002

// Background is a part of the master that runs next to the http server.
// Run returns at once and the part stops when ctx is done.
type Background interface {
	Run(ctx context.Context)
}

// RunMaster runs the device manager and the background parts and serves http
// until the server fails. Everything started stops when it returns or ctx is
// done.
func (s *Service) RunMaster(ctx context.Context, background ...Background) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := s.deviceManager.Run(ctx, true); err != nil {
		return err
	}
	defer s.deviceManager.StopObserver()

	for _, part := range background {
		part.Run(ctx)
	}

	s.publicRouter.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	if s.cfg.Port == 0 {
		s.cfg.Port = defaultMasterPort
	}

	err := s.publicRouter.Run(fmt.Sprintf(":%d", s.cfg.Port))
	logrus.Infof("Stopping Server")
	return err
}
