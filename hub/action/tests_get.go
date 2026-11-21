package action

import "github.com/fsuhrau/automationhub/device/appproto"

type TestsGet struct {
	Tests   []appproto.Test
	Success bool
}

func (a *TestsGet) GetActionType() appproto.ActionType {
	return appproto.ActionType_GetTests
}

func (a *TestsGet) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_GetTests, appproto.RequestData{})
}

func (a *TestsGet) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	tests, err := payloadValue(response.Payload.Tests, "tests")
	if err != nil {
		return err
	}
	a.Tests = tests.Tests
	return nil
}
