package action

type TouchElement struct {
	ElementID string
	Success   bool
}

func (a *TouchElement) GetActionType() ActionType {
	return ActionType_ElementTouch
}

func (a *TouchElement) Serialize() ([]byte, error) {
	return serialize(ActionType_ElementTouch, RequestData{Id: &a.ElementID})
}

func (a *TouchElement) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
