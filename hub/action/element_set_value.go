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

func (a *SetValue) GetActionType() ActionType {
	return ActionType_ElementSetValue
}

func (a *SetValue) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementSetValue,
		Payload: &Request_SetAttr{&SetAttr{Id: a.ElementID, Attr: a.Attr, Val:a.Value}},
	}
	return proto.Marshal(req)
}

func (a *SetValue) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
