package action

import "github.com/fsuhrau/automationhub/device/appproto"

type TestStart struct {
	Assembly string
	Class    string
	Method   string
	Env      map[string]string
	Success  bool
}

func (a *TestStart) GetActionType() appproto.ActionType {
	return appproto.ActionType_ExecuteTest
}

func (a *TestStart) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_ExecuteTest, appproto.RequestData{Test: &appproto.Test{
		Assembly:  a.Assembly,
		Class:     a.Class,
		Method:    a.Method,
		Parameter: a.Env,
	}})
}

func (a *TestStart) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	return nil
}
