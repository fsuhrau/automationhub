package action

import (
	"google.golang.org/protobuf/proto"
)

type TestsGet struct {
	Tests   []*Test
	Success bool
}

func (a *TestsGet) GetActionType() ActionType {
	return ActionType_GetTests
}

func (a *TestsGet) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_GetTests,
	}
	return proto.Marshal(req)
}

func (a *TestsGet) ProcessResponse(response *Response) error {
	a.Tests = response.GetTests().Tests
	a.Success = response.Success
	return nil
}
