package action

import (
	"encoding/json"
)

type DragAndDrop struct {
	From string
	To   string
}

func (a *DragAndDrop) GetActionType() ActionType {
	return ActionType_DragAndDrop
}

func (a *DragAndDrop) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_DragAndDrop,
		Payload:    RequestData{MoveElement: &MoveElement{From: a.From, To: a.To}},
	}
	return json.Marshal(req)
}

func (a *DragAndDrop) ProcessResponse(response *Response) error {
	return nil
}
