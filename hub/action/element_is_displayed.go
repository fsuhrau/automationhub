package action

type IsDisplayed struct {
	ElementID   string
	Success     bool
	IsDisplayed bool
}

func (a *IsDisplayed) GetActionType() ActionType {
	return ActionType_ElementIsDisplayed
}

func (a *IsDisplayed) Serialize() ([]byte, error) {
	return serialize(ActionType_ElementIsDisplayed, RequestData{Id: &a.ElementID})
}

func (a *IsDisplayed) ProcessResponse(response *Response) error {
	a.Success = response.Success
	visible, err := payloadValue(response.Payload.Visible, "visible")
	if err != nil {
		return err
	}
	a.IsDisplayed = visible
	return nil
}
