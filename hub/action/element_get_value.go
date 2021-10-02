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

func (a *GetValue) GetActionType() ActionType {
	return ActionType_ElementGetValue
}

func (a *GetValue) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementGetValue,
		Payload:    &Request_GetAttr{&GetAttr{Id: a.ElementID, Attr: a.Attr}},
	}
	return proto.Marshal(req)
}

func (a *GetValue) ProcessResponse(response *Response) error {
	a.Value = response.GetValue()
	a.Success = response.Success
	return nil
}
