package action

import (
	"google.golang.org/protobuf/proto"
)

type UnityReset struct {
	Assembly string
	Class string
	Method string
	Success bool
}

func (a *UnityReset) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_UnityReset,
	}
	return proto.Marshal(req)
}

func (a *UnityReset) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Success = resp.Success
	return nil
}
