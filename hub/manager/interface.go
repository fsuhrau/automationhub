package manager

import (
	"context"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
)

type NodeIdentifier string

type Sessions interface {
	Run(ctx context.Context)
	GetSession(sessionID string) (Session, error)
	CreateNewSession(logger *logrus.Entry, properties *device.Properties, parameter *app.Parameter) Session
	AddSession(s Session)
	StopSession(s Session) error
	GetSessions() []Session
	GetSessionDetails(sessionID string) Session
}

type Storage interface {
	StoreSceneGraph(data []byte) (string, error)
	StoreImage(content []byte) (string, error)
	Close() error
}
