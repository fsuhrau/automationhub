package action

import (
	"google.golang.org/protobuf/proto"
)

type Scroll struct {
	OffsetX int64
	OffsetY int64
}

func (a *Scroll) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_Move,
		Payload: &Request_MoveOffset{&MoveOffset{From: a.OffsetX, To:a.OffsetY}},
	}
	return proto.Marshal(req)
}

func (a *Scroll) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	return nil
}
