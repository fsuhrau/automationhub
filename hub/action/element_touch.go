package action

import (
	"google.golang.org/protobuf/proto"
)

type TouchElement struct {
	ElementID string
	Success bool
}

func (a *TouchElement) GetActionType() ActionType {
	return ActionType_ElementTouch
}

func (a *TouchElement) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementTouch,
		Payload: &Request_Id{Id: a.ElementID},
	}
	return proto.Marshal(req)
}

func (a *TouchElement) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
