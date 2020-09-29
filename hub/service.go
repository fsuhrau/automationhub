package hub

import (
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type Selectable struct {
	ID   string
	Name string
}
type Selectables []*Selectable

func (s Selectables) Len() int      { return len(s) }
func (s Selectables) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByName struct{ Selectables }

func (s ByName) Less(i, j int) bool { return s.Selectables[i].Name < s.Selectables[j].Name }

type Service struct {
	server         *http.Server
	deviceManager  *DeviceManager
	sessionManager *SessionManager
	logger         *logrus.Logger
	hostIP         net.IP
	// sessions       map[string]*Session
}

func NewService() *Service {
	logger := logrus.New()
	level, err := logrus.ParseLevel(viper.GetString("log"))
	if err != nil {
		logrus.Infof("Parse Log Level: %s", err)
		logrus.Info("using default info")
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(level)
	}
	logger.Formatter = new(prefixed.TextFormatter)
	deviceManager := NewManager(logger)
	sessionManager := NewSessionManager(logger, deviceManager)
	return &Service{logger: logger, sessionManager: sessionManager, deviceManager: deviceManager}
}
