package action

type Scroll struct {
	OffsetX int64
	OffsetY int64
}

func (a *Scroll) GetActionType() ActionType {
	return ActionType_Move
}

func (a *Scroll) Serialize() ([]byte, error) {
	return serialize(ActionType_Move, RequestData{MoveOffset: &MoveOffset{From: a.OffsetX, To: a.OffsetY}})
}

func (a *Scroll) ProcessResponse(response *Response) error {
	return nil
}
