package action

import (
	"google.golang.org/protobuf/proto"
)

type Custom struct {
	RequestData string
	Content []byte
	Success bool
}

func (a *Custom) GetActionType() ActionType {
	return ActionType_Custom
}

func (a *Custom) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_Custom,
		Payload: &Request_Data{Data: a.RequestData},
	}
	return proto.Marshal(req)
}

func (a *Custom) ProcessResponse(response *Response) error {
	a.Content = response.GetData()
	a.Success = response.Success
	return nil
}
