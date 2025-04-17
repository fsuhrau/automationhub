package action

import (
	"encoding/json"
)

type GetValue struct {
	ElementID string
	Success   bool
	Attr      string
	Value     string
}

func (a *GetValue) GetActionType() ActionType {
	return ActionType_ElementGetValue
}

func (a *GetValue) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementGetValue,
		Payload:    RequestData{GetAttr: &GetAttr{Id: a.ElementID, Attr: a.Attr}},
	}
	return json.Marshal(req)
}

func (a *GetValue) ProcessResponse(response *Response) error {
	a.Value = *response.Payload.Value
	a.Success = response.Success
	return nil
}
