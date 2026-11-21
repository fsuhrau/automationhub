package action

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
	return serialize(ActionType_ElementGetValue, RequestData{GetAttr: &GetAttr{Id: a.ElementID, Attr: a.Attr}})
}

func (a *GetValue) ProcessResponse(response *Response) error {
	a.Success = response.Success
	value, err := payloadValue(response.Payload.Value, "value")
	if err != nil {
		return err
	}
	a.Value = value
	return nil
}
