package node

import (
	"github.com/fsuhrau/automationhub/config/protocol"
	"github.com/fsuhrau/automationhub/hub/node/jsonrpc"
	"github.com/gorilla/websocket"
	"github.com/matishsiao/goInfo"
	"net/http"
	"net/rpc"
	"net/url"
	"os"
	"time"
)

var rpcServer = rpc.DefaultServer

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

	headers := http.Header{}
	if s.cfg.Auth.Token != nil {
		headers.Set("X-Auth-Token", s.cfg.Auth.Token.AuthToken)
	}

	dialer := websocket.Dialer{
		// Proxy:             http.ProxyURL(proxyURL),
		HandshakeTimeout:  45 * time.Second,
		ReadBufferSize:    protocol.SocketFrameSize,
		WriteBufferSize:   protocol.SocketFrameSize,
		EnableCompression: true,
	}
	c, _, err := dialer.Dial(serverURL.String(), headers)
	if err != nil {
		s.logger.Errorf("connect to master failed: %v", err)
		return
	}

	gi, _ := goInfo.GetInfo()
	if err := s.registerNode(c, s.cfg.Identifier, gi.Hostname, gi.OS, os.Getenv("PATH"), s.cfg.Port, s.managers); err != nil {
		s.logger.Errorf("register node failed: %v", err)
		_ = c.Close()
		return
	}

	s.connMu.Lock()
	s.c = c
	s.connMu.Unlock()
	dialDone = true
	s.dialing.Store(false)

	rpcServer.ServeCodec(jsonrpc.NewServerCodec(c))

	s.connMu.Lock()
	if s.c == c {
		s.c = nil
	}
	s.connMu.Unlock()
}

func (s *Service) closeConnection() {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	if s.c != nil {
		_ = s.c.Close()
		s.c = nil
	}
}
