package action

import (
	"encoding/json"
)

type Custom struct {
	RequestData string
	Content     []byte
	Success     bool
}

func (a *Custom) GetActionType() ActionType {
	return ActionType_Custom
}

func (a *Custom) Serialize() ([]byte, error) {
	data := []byte(a.RequestData)
	req := &Request{
		ActionType: ActionType_Custom,
		Payload:    RequestData{Data: &data},
	}
	return json.Marshal(req)
}

func (a *Custom) ProcessResponse(response *Response) error {
	a.Content = *response.Payload.Data
	a.Success = response.Success
	return nil
}
