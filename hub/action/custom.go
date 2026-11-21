package action

type Custom struct {
	RequestData string
	Content     []byte
	Success     bool
}

func (a *Custom) GetActionType() ActionType {
	return ActionType_Custom
}

func (a *Custom) Serialize() ([]byte, error) {
	data := []byte(a.RequestData)
	return serialize(ActionType_Custom, RequestData{Data: &data})
}

func (a *Custom) ProcessResponse(response *Response) error {
	a.Success = response.Success
	content, err := payloadValue(response.Payload.Data, "data")
	if err != nil {
		return err
	}
	a.Content = content
	return nil
}
