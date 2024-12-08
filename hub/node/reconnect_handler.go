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
	lastPing time.Time
	cfg      config.Service
	managers []string
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
	/*
		serverURL := url.URL{Scheme: "ws", Host: h.cfg.MasterURL, Path: "/node/connect"}

		c, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
		if err != nil {
			log.Fatal("dial: ", err)
		}

		gi, _ := goInfo.GetInfo()
		if err := s.RegisterNode(s.cfg.Identifier, gi.Hostname, gi.OS, os.Getenv("PATH"), s.cfg.Port, s.managers); err != nil {
			fmt.Printf("%v\n", err)
			return
		}
	*/
	return nil
}
