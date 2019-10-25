package action

import (
	"google.golang.org/protobuf/proto"
)

type LongTouchElement struct {
	ElementID string
	Success bool
}

func (a *LongTouchElement) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_LongTouch,
		Payload: &Request_Id{Id: a.ElementID},
	}
	return proto.Marshal(req)
}

func (a *LongTouchElement) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Success = resp.Success
	return nil
}
