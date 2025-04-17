package node

import (
	"context"
	"github.com/fsuhrau/automationhub/config"
	"time"
)

type ConnectionHandler interface {
	PingReceived()
}

type ReconnectHandler struct {
	lastPing        time.Time
	cfg             config.Service
	managers        []string
	ConnectFunction func()
}

func (h *ReconnectHandler) PingReceived() {
	h.lastPing = time.Now()
}

func (h *ReconnectHandler) ObserveConnectionState(ctx context.Context, cfg config.Service, managers []string) {
	h.cfg = cfg
	h.managers = managers
	time.Sleep(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
		default:
			if h.lastPing.Add(30 * time.Second).Before(time.Now()) {
				_ = h.HandleConnect()
			}
			time.Sleep(5 * time.Second)
		}
	}
}

func (h *ReconnectHandler) HandleConnect() error {
	go h.ConnectFunction()
	return nil
}
