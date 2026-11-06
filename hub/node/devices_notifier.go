package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DevicesNotifierDelay coalesces bursts of device events (a phone that is
// plugged in usually produces several state changes) into one notification.
const DevicesNotifierDelay = 300 * time.Millisecond

// DevicesNotifier tells the master that this node's device list changed.
type DevicesNotifier struct {
	masterURL string
	node      string
	authToken *string
	client    *http.Client

	mu      sync.Mutex
	pending *time.Timer
}

// NewDevicesNotifier returns nil when there is no master to notify (the
// service runs as master itself); Notify on nil is a no-op.
func NewDevicesNotifier(masterURL, nodeIdentifier string, authToken *string) *DevicesNotifier {
	if masterURL == "" || nodeIdentifier == "" {
		return nil
	}
	return &DevicesNotifier{
		masterURL: masterURL,
		node:      nodeIdentifier,
		authToken: authToken,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify schedules a notification; calls within DevicesNotifierDelay are merged.
func (n *DevicesNotifier) Notify() {
	if n == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.pending != nil {
		return
	}
	n.pending = time.AfterFunc(DevicesNotifierDelay, func() {
		n.mu.Lock()
		n.pending = nil
		n.mu.Unlock()
		if err := n.send(); err != nil {
			logrus.Debugf("notify master about device change: %v", err)
		}
	})
}

func (n *DevicesNotifier) send() error {
	body, _ := json.Marshal(map[string]string{"node": n.node})
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
	if resp.StatusCode >= 300 {
		return fmt.Errorf("master answered %s", resp.Status)
	}
	return nil
}
