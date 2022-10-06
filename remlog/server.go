package remlog

import (
	"bytes"
	"fmt"
	"io"
	"net"

	"github.com/sirupsen/logrus"
)

type Service struct {
	conn *net.UDPConn
	log  *logrus.Entry
}

func NewService(logger *logrus.Logger) *Service {
	return &Service{
		log: logger.WithFields(logrus.Fields{
			"prefix": "remlog",
		}),
	}
}

func (s *Service) Run(logLine func(string)) error {
	add := "0.0.0.0:6666"
	s.log.Infof("starting service on: %s", add)

	// resolve udp address
	serverAddr, err := net.ResolveUDPAddr("", add)
	if err != nil {
		return err
	}

	// create udo server
	s.conn, err = net.ListenUDP("udp", serverAddr)
	if err != nil {
		return err
	}
	go func() {
		defer s.conn.Close()
		// read buffer
		buffer := make([]byte, 2048)
		for {
			n, address, err := s.conn.ReadFromUDP(buffer)
			connectionLog := s.log.WithFields(logrus.Fields{
				"address": fmt.Sprintf("%v", address),
			})

			if err != nil {
				connectionLog.Errorf("error receive %v", err)
			}

			z := bytes.NewReader(buffer[0:n])
			b, err := io.ReadAll(z)
			if err != nil {
				connectionLog.Errorf("error read %v", err)
			}
			logLine(string(b))
		}

	}()
	return nil
}
