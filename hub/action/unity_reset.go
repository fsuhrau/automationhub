package action

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
	return serialize(ActionType_UnityReset, RequestData{})
}

func (a *UnityReset) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
