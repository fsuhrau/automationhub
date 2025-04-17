package action

import (
	"encoding/json"
)

type IsDisplayed struct {
	ElementID   string
	Success     bool
	IsDisplayed bool
}

func (a *IsDisplayed) GetActionType() ActionType {
	return ActionType_ElementIsDisplayed
}

func (a *IsDisplayed) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementIsDisplayed,
		Payload:    RequestData{Id: &a.ElementID},
	}
	return json.Marshal(req)
}

func (a *IsDisplayed) ProcessResponse(response *Response) error {
	a.Success = response.Success
	a.IsDisplayed = *response.Payload.Visible
	return nil
}
