package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/hub/action"
	"net/http"
	"time"
)

type RemoteActionHandler struct {
	masterURL string
	node      string
	deviceId  string
}

type request struct {
	Node     string           `json:"node"`
	DeviceID string           `json:"device_id"`
	Response *action.Response `json:"response"`
}

func NewRemoteActionHandler(masterURL, nodeIdentifier, deviceId string) *RemoteActionHandler {
	return &RemoteActionHandler{masterURL: masterURL, node: nodeIdentifier, deviceId: deviceId}
}

func (w *RemoteActionHandler) OnActionResponse(d interface{}, response *action.Response) {
	reqData, err := json.Marshal(&request{
		DeviceID: w.deviceId,
		Node:     w.node,
		Response: response,
	})
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/node/action", w.masterURL), bytes.NewBuffer(reqData))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	return
}
