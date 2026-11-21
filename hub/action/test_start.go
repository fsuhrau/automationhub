package action

type TestStart struct {
	Assembly string
	Class    string
	Method   string
	Env      map[string]string
	Success  bool
}

func (a *TestStart) GetActionType() ActionType {
	return ActionType_ExecuteTest
}

func (a *TestStart) Serialize() ([]byte, error) {
	return serialize(ActionType_ExecuteTest, RequestData{Test: &Test{
		Assembly:  a.Assembly,
		Class:     a.Class,
		Method:    a.Method,
		Parameter: a.Env,
	}})
}

func (a *TestStart) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
