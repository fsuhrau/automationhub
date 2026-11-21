package action

type TestsGet struct {
	Tests   []Test
	Success bool
}

func (a *TestsGet) GetActionType() ActionType {
	return ActionType_GetTests
}

func (a *TestsGet) Serialize() ([]byte, error) {
	return serialize(ActionType_GetTests, RequestData{})
}

func (a *TestsGet) ProcessResponse(response *Response) error {
	a.Success = response.Success
	tests, err := payloadValue(response.Payload.Tests, "tests")
	if err != nil {
		return err
	}
	a.Tests = tests.Tests
	return nil
}
