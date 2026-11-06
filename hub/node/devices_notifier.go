package node

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DevicesNotifierDelay coalesces bursts of device events (a phone that is
// plugged in usually produces several state changes) into one notification.
const DevicesNotifierDelay = 300 * time.Millisecond

// Retry bounds for a device list the master didn't accept.
const (
	devicesNotifierMinRetry = 1 * time.Second
	devicesNotifierMaxRetry = 30 * time.Second
)

// DevicesChangedRequest is the body of POST /node/devices. Devices is the
// node's complete device list, so the master applies it without asking the
// node again. A body without devices (older nodes) makes the master fetch
// the list over RPC.
type DevicesChangedRequest struct {
	Node string `json:"node" binding:"required"`
	// SentAt (unix nano, node clock) orders pushes; the master drops one
	// that is older than the last it applied.
	SentAt  int64              `json:"sentAt,omitempty"`
	Devices *[]*DeviceResponse `json:"devices,omitempty"`
}

// errTokenRejected is a push the master refused for its X-Auth-Token.
var errTokenRejected = errors.New("master rejected the node token")

// DevicesNotifier pushes this node's device list to the master whenever it
// changed. Notify is cheap: the list is only sent when it differs from the
// last one the master accepted, and a failed send is retried with backoff.
type DevicesNotifier struct {
	masterURL string
	node      string
	authToken *string
	client    *http.Client
	snapshot  func() []*DeviceResponse

	sendMu   sync.Mutex // serializes flushes, guards lastSent and rejected
	lastSent []byte
	rejected bool // the master refused the token; warned once until a push works

	mu      sync.Mutex // guards pending and retry
	pending *time.Timer
	retry   time.Duration
}

// NewDevicesNotifier returns nil when there is no master to notify (the
// service runs as master itself); Notify on nil is a no-op. snapshot returns
// the node's current device list.
func NewDevicesNotifier(masterURL, nodeIdentifier string, authToken *string, snapshot func() []*DeviceResponse) *DevicesNotifier {
	if masterURL == "" || nodeIdentifier == "" || snapshot == nil {
		return nil
	}
	return &DevicesNotifier{
		masterURL: masterURL,
		node:      nodeIdentifier,
		authToken: authToken,
		snapshot:  snapshot,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify schedules a push; calls within DevicesNotifierDelay are merged.
func (n *DevicesNotifier) Notify() {
	if n == nil {
		return
	}
	n.schedule(DevicesNotifierDelay)
}

func (n *DevicesNotifier) schedule(delay time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.pending != nil {
		return
	}
	n.pending = time.AfterFunc(delay, n.flush)
}

func (n *DevicesNotifier) flush() {
	n.mu.Lock()
	n.pending = nil
	n.mu.Unlock()

	n.sendMu.Lock()
	defer n.sendMu.Unlock()

	devices := n.snapshot()
	if devices == nil {
		devices = []*DeviceResponse{}
	}
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].Manager != devices[j].Manager {
			return devices[i].Manager < devices[j].Manager
		}
		return devices[i].DeviceID < devices[j].DeviceID
	})
	data, err := json.Marshal(devices)
	if err != nil {
		logrus.Errorf("encode device list: %v", err)
		return
	}
	if bytes.Equal(data, n.lastSent) {
		return
	}

	if err := n.send(devices); err != nil {
		n.mu.Lock()
		n.retry = min(max(n.retry*2, devicesNotifierMinRetry), devicesNotifierMaxRetry)
		retry := n.retry
		n.mu.Unlock()
		if errors.Is(err, errTokenRejected) {
			// retrying won't help until the token is replaced, so say it once
			// where the operator sees it instead of at debug level
			if !n.rejected {
				n.rejected = true
				logrus.Warnf("push device list to master: %v; issue a new token for this node in the master's settings and put it in the node config (retry in %v)", err, retry)
			}
		} else {
			logrus.Debugf("push device list to master (retry in %v): %v", retry, err)
		}
		n.schedule(retry)
		return
	}
	n.rejected = false
	n.lastSent = data
	n.mu.Lock()
	n.retry = 0
	n.mu.Unlock()
}

func (n *DevicesNotifier) send(devices []*DeviceResponse) error {
	body, err := json.Marshal(DevicesChangedRequest{Node: n.node, SentAt: time.Now().UnixNano(), Devices: &devices})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/node/devices", n.masterURL), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if n.authToken != nil {
		req.Header.Set("X-Auth-Token", *n.authToken)
	}
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%w (%s)", errTokenRejected, resp.Status)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("master answered %s", resp.Status)
	}
	return nil
}
