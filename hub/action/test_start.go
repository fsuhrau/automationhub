package action

import (
	"google.golang.org/protobuf/proto"
)

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
	req := &Request{
		ActionType: ActionType_ExecuteTest,
		Payload: &Request_Test{&Test{
			Assembly:  a.Assembly,
			Class:     a.Class,
			Method:    a.Method,
			Parameter: a.Env,
		}},
	}
	return proto.Marshal(req)
}

func (a *TestStart) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
