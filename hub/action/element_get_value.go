package action

import (
	"google.golang.org/protobuf/proto"
)

type GetValue struct {
	ElementID string
	Success   bool
	Attr      string
	Value     string
}

func (a *GetValue) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementGetValue,
		Payload:    &Request_GetAttr{&GetAttr{Id: a.ElementID, Attr: a.Attr}},
	}
	return proto.Marshal(req)
}

func (a *GetValue) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Value = resp.GetValue()
	a.Success = resp.Success
	return nil
}
