package node

import (
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/matishsiao/goInfo"
)

func (s *Service) authHeaders() http.Header {
	headers := http.Header{}
	if s.cfg.Auth.Token != nil {
		headers.Set("X-Auth-Token", s.cfg.Auth.Token.AuthToken)
	}
	return headers
}

// announce sends the registration request describing this node to the master.
func (s *Service) announce(c *websocket.Conn) error {
	gi, _ := goInfo.GetInfo()
	return s.registerNode(c, nodeInfo{
		identifier: s.cfg.Identifier,
		hostname:   gi.Hostname,
		os:         gi.OS,
		path:       os.Getenv("PATH"),
		port:       s.cfg.Port,
		managers:   s.managers,
	})
}

func (s *Service) setConn(c *websocket.Conn) {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	s.c = c
}

func (s *Service) clearConn(c *websocket.Conn) {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	if s.c == c {
		s.c = nil
	}
}
