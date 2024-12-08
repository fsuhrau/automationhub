package action

import "encoding/json"

type UnityReset struct {
	Assembly string
	Class    string
	Method   string
	Success  bool
}

func (a *UnityReset) GetActionType() ActionType {
	return ActionType_UnityReset
}

func (a *UnityReset) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_UnityReset,
	}
	return json.Marshal(req)
}

func (a *UnityReset) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
