package node

import (
	"github.com/fsuhrau/automationhub/config/protocol"
	"github.com/fsuhrau/automationhub/hub/node/jsonrpc"
	"github.com/gorilla/websocket"
	"net/url"
	"time"
)

func (s *Service) ConnectAndServe() {
	// only one dial at a time; a running ServeCodec of an older connection is
	// terminated by closing its socket below.
	if !s.dialing.CompareAndSwap(false, true) {
		return
	}
	dialDone := false
	defer func() {
		if !dialDone {
			s.dialing.Store(false)
		}
	}()

	s.closeConnection()

	serverURL := url.URL{Scheme: "ws", Host: s.cfg.MasterURL, Path: "/node/connect"}

	// proxyURL, _ := url.Parse("http://10.35.111.236:8888")

	dialer := websocket.Dialer{
		// Proxy:             http.ProxyURL(proxyURL),
		HandshakeTimeout:  45 * time.Second,
		ReadBufferSize:    protocol.SocketFrameSize,
		WriteBufferSize:   protocol.SocketFrameSize,
		EnableCompression: true,
	}
	c, _, err := dialer.Dial(serverURL.String(), s.authHeaders())
	if err != nil {
		s.logger.Errorf("connect to master failed: %v", err)
		return
	}

	if err := s.announce(c); err != nil {
		s.logger.Errorf("register node failed: %v", err)
		_ = c.Close()
		return
	}

	s.setConn(c)
	dialDone = true
	s.dialing.Store(false)

	s.rpcServer.ServeCodec(jsonrpc.NewServerCodec(c))

	s.clearConn(c)
}

func (s *Service) closeConnection() {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	if s.c != nil {
		_ = s.c.Close()
		s.c = nil
	}
}
