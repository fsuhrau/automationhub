package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/device/appproto"
	"net/http"
	"time"
)

type ActionHandler struct {
	masterURL string
	node      string
	deviceId  string
	authToken *string
}

type request struct {
	Node     string             `json:"node"`
	DeviceID string             `json:"deviceId"`
	Response *appproto.Response `json:"response"`
}

//nolint:revive // argument-limit: signature is fixed by the platform handlers that call it through hub/node
func NewActionHandler(masterURL, nodeIdentifier, deviceId string, authToken *string) *ActionHandler {
	return &ActionHandler{masterURL: masterURL, node: nodeIdentifier, deviceId: deviceId, authToken: authToken}
}

func (w *ActionHandler) OnActionResponse(d interface{}, response *appproto.Response) {
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
	if w.authToken != nil {
		req.Header.Set("X-Auth-Token", *w.authToken)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return
	}
}
