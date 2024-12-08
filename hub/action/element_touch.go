package action

import (
	"encoding/json"
)

type TouchElement struct {
	ElementID string
	Success   bool
}

func (a *TouchElement) GetActionType() ActionType {
	return ActionType_ElementTouch
}

func (a *TouchElement) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementTouch,
		Payload:    RequestData{Id: &a.ElementID},
	}
	return json.Marshal(req)
}

func (a *TouchElement) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
