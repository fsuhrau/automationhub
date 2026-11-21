package action

import (
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/device/appproto"
)

// serialize builds the wire request of an action: its type and the payload
// that carries the action's specific fields.
func serialize(actionType appproto.ActionType, payload appproto.RequestData) ([]byte, error) {
	return json.Marshal(&appproto.Request{ActionType: actionType, Payload: payload})
}

// payloadValue returns the value a response payload field points to. A
// response without that field is malformed and yields an error.
func payloadValue[T any](field *T, name string) (T, error) {
	if field == nil {
		var zero T
		return zero, fmt.Errorf("response payload has no %s", name)
	}
	return *field, nil
}

// touchPayload is the payload shared by all touch position actions.
func touchPayload(x, y int64) appproto.RequestData {
	return appproto.RequestData{Touch: &appproto.Touch{Xoffset: x, Yoffset: y}}
}
