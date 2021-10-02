package action

import (
	"google.golang.org/protobuf/proto"
)

type IsDisplayed struct {
	ElementID string
	Success bool
	IsDisplayed bool
}

func (a *IsDisplayed) GetActionType() ActionType {
	return ActionType_ElementIsDisplayed
}

func (a *IsDisplayed) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementIsDisplayed,
		Payload: &Request_Id{Id: a.ElementID},
	}
	return proto.Marshal(req)
}

func (a *IsDisplayed) ProcessResponse(response *Response) error {
	a.Success = response.Success
	a.IsDisplayed = response.GetVisible()
	return nil
}

