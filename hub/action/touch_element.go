package action

import (
	"google.golang.org/protobuf/proto"
)

type TouchElement struct {
	ElementID string
	Success bool
}

func (a *TouchElement) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementTouch,
		Payload: &Request_Id{Id: a.ElementID},
	}
	return proto.Marshal(req)
}

func (a *TouchElement) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Success = resp.Success
	return nil
}
