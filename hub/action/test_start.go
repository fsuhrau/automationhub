package action

import (
	"google.golang.org/protobuf/proto"
)

type TestStart struct {
	Assembly string
	Class string
	Method string
	Success bool
}

func (a *TestStart) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ExecuteTest,
		Payload: &Request_Test{&Test{
			Assembly: a.Assembly,
			Class: a.Class,
			Method: a.Method,
		}},
	}
	return proto.Marshal(req)
}

func (a *TestStart) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Success = resp.Success
	return nil
}
