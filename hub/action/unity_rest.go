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

func (a *UnityReset) GetActionType() ActionType {
	return ActionType_UnityReset
}

func (a *UnityReset) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_UnityReset,
	}
	return proto.Marshal(req)
}

func (a *UnityReset) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
