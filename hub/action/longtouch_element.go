package action

import (
	"google.golang.org/protobuf/proto"
)

type LongTouchElement struct {
	ElementID string
	Success bool
}

func (a *LongTouchElement) GetActionType() ActionType {
	return ActionType_LongTouch
}

func (a *LongTouchElement) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_LongTouch,
		Payload: &Request_Id{Id: a.ElementID},
	}
	return proto.Marshal(req)
}

func (a *LongTouchElement) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
