package action

import "encoding/json"

type TestsGet struct {
	Tests   []Test
	Success bool
}

func (a *TestsGet) GetActionType() ActionType {
	return ActionType_GetTests
}

func (a *TestsGet) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_GetTests,
	}
	return json.Marshal(req)
}

func (a *TestsGet) ProcessResponse(response *Response) error {
	a.Tests = response.Payload.Tests.Tests
	a.Success = response.Success
	return nil
}
