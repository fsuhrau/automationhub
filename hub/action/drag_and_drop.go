package action

type DragAndDrop struct {
	From string
	To   string
}

func (a *DragAndDrop) GetActionType() ActionType {
	return ActionType_DragAndDrop
}

func (a *DragAndDrop) Serialize() ([]byte, error) {
	return serialize(ActionType_DragAndDrop, RequestData{MoveElement: &MoveElement{From: a.From, To: a.To}})
}

func (a *DragAndDrop) ProcessResponse(response *Response) error {
	return nil
}
