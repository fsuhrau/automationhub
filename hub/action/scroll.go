package action

import (
	"google.golang.org/protobuf/proto"
)

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
		Payload: &Request_MoveOffset{&MoveOffset{From: a.OffsetX, To:a.OffsetY}},
	}
	return proto.Marshal(req)
}

func (a *Scroll) ProcessResponse(response *Response) error {
	return nil
}
