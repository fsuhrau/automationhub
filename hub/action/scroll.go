package action

import "encoding/json"

type Scroll struct {
	OffsetX int64
	OffsetY int64
}

func (a *Scroll) GetActionType() ActionType {
	return ActionType_Move
}

func (a *Scroll) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_Move,
		Payload:    RequestData{MoveOffset: &MoveOffset{From: a.OffsetX, To: a.OffsetY}},
	}
	return json.Marshal(req)
}

func (a *Scroll) ProcessResponse(response *Response) error {
	return nil
}
