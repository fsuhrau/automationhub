package node

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/fsuhrau/automationhub/config"
)

const (
	pingTimeout        = 30 * time.Second
	observeInterval    = 5 * time.Second
	initialObserveWait = 1 * time.Second
)

type ConnectionHandler interface {
	PingReceived()
}

type ReconnectHandler struct {
	lastPing        atomic.Int64 // unix nano of last ping (or connect attempt)
	cfg             config.Service
	managers        []string
	ConnectFunction func()
}

func (h *ReconnectHandler) PingReceived() {
	h.lastPing.Store(time.Now().UnixNano())
}

func (h *ReconnectHandler) pingExpired(now time.Time) bool {
	return time.Unix(0, h.lastPing.Load()).Add(pingTimeout).Before(now)
}

func (h *ReconnectHandler) ObserveConnectionState(ctx context.Context, cfg config.Service, managers []string) {
	h.cfg = cfg
	h.managers = managers

	select {
	case <-ctx.Done():
		return
	case <-time.After(initialObserveWait):
	}

	ticker := time.NewTicker(observeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if h.pingExpired(now) {
				_ = h.HandleConnect()
			}
		}
	}
}

// HandleConnect starts a connection attempt in the background. The ping timer
// is reset so the observer gives the new connection a full ping interval
// before trying again. ConnectFunction must close a previous connection and
// ignore calls while a dial is still in progress.
func (h *ReconnectHandler) HandleConnect() error {
	h.PingReceived()
	go h.ConnectFunction()
	return nil
}
