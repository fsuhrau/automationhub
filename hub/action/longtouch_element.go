package action

import "encoding/json"

type LongTouchElement struct {
	ElementID string
	Success   bool
}

func (a *LongTouchElement) GetActionType() ActionType {
	return ActionType_LongTouch
}

func (a *LongTouchElement) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_LongTouch,
		Payload:    RequestData{Id: &a.ElementID},
	}
	return json.Marshal(req)
}

func (a *LongTouchElement) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
