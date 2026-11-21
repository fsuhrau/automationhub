package action

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
	return serialize(ActionType_ElementSetValue, RequestData{SetAttr: &SetAttr{Id: a.ElementID, Attr: a.Attr, Val: a.Value}})
}

func (a *SetValue) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
