package hub

import (
	"net/http"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/config/protocol"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	DeviceConnectionTimeout = 5 * time.Minute

	deviceRefreshInterval = 5 * time.Second
)

// DeviceManager owns the device handlers, the websocket the apps connect
// through and the dispatch of the responses the apps send back. The parts live
// in the device_manager_*.go files.
type DeviceManager struct {
	deviceHandlers map[string]device.Handler
	stop           bool
	log            *logrus.Entry
	upgrader       websocket.Upgrader
	masterUrl      string
	nodeIdentifier string
	authToken      *string

	refreshSignal   chan struct{}
	refreshInterval time.Duration

	// pushes device changes to the master; a no-op on the master itself
	notifier        ChangeNotifier
	newNotifier     NotifierFactory
	lockEvents      *events.Bus[events.DeviceLockChangedPayload]
	responseHandler map[action.ActionType]responseHandler
}

// DeviceManagerOption replaces a default dependency of the DeviceManager.
type DeviceManagerOption func(*DeviceManager)

// WithNotifierFactory replaces the factory that builds the notifier in Run.
func WithNotifierFactory(factory NotifierFactory) DeviceManagerOption {
	return func(dm *DeviceManager) { dm.newNotifier = factory }
}

// WithLockEvents replaces the bus that announces lock changes.
func WithLockEvents(bus *events.Bus[events.DeviceLockChangedPayload]) DeviceManagerOption {
	return func(dm *DeviceManager) { dm.lockEvents = bus }
}

func NewDeviceManager(logger *logrus.Logger, service config.Service, opts ...DeviceManagerOption) *DeviceManager {
	var authToken *string
	if service.Auth.Token != nil {
		authToken = &service.Auth.Token.AuthToken
	}
	dm := &DeviceManager{
		log:             logger.WithFields(logrus.Fields{"prefix": "dm"}),
		masterUrl:       service.MasterURL,
		nodeIdentifier:  service.Identifier,
		authToken:       authToken,
		deviceHandlers:  make(map[string]device.Handler),
		upgrader:        newUpgrader(service.Cors),
		refreshSignal:   make(chan struct{}, 1),
		refreshInterval: deviceRefreshInterval,
		notifier:        noopNotifier{},
		lockEvents:      events.DeviceLockChanged,
	}
	dm.newNotifier = dm.masterNotifierFactory
	dm.responseHandler = dm.responseHandlers()
	for _, opt := range opts {
		opt(dm)
	}
	return dm
}

func newUpgrader(cors []string) websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  protocol.SocketFrameSize,
		WriteBufferSize: protocol.SocketFrameSize,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			// mobile has no origin find a better check
			return origin == "" || originAllowed(cors, origin)
		},
	}
}

func originAllowed(allowed []string, origin string) bool {
	for i := range allowed {
		if allowed[i] == origin {
			return true
		}
	}
	return false
}

func (dm *DeviceManager) StopObserver() {
	dm.stop = true
}
