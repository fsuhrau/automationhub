package action

import (
	"google.golang.org/protobuf/proto"
)

type SetValue struct {
	ElementID string
	Success bool
	Attr string
	Value string
}

func (a *SetValue) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementSetValue,
		Payload: &Request_SetAttr{&SetAttr{Id: a.ElementID, Attr: a.Attr, Val:a.Value}},
	}
	return proto.Marshal(req)
}

func (a *SetValue) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Success = resp.Success
	return nil
}
