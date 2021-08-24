package action

import (
	"google.golang.org/protobuf/proto"
)

type TestStart struct {
	TestClass string
	TestMethod string
	Success bool
}

func (a *TestStart) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ExecuteTest,
		Payload: &Request_Test{&Test{
			Type: a.TestClass,
			Method: a.TestMethod,
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
