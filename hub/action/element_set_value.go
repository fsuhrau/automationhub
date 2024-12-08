package action

import "encoding/json"

type SetValue struct {
	ElementID string
	Success   bool
	Attr      string
	Value     string
}

func (a *SetValue) GetActionType() ActionType {
	return ActionType_ElementSetValue
}

func (a *SetValue) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementSetValue,
		Payload:    RequestData{SetAttr: &SetAttr{Id: a.ElementID, Attr: a.Attr, Val: a.Value}},
	}
	return json.Marshal(req)
}

func (a *SetValue) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
