package action

import (
	"google.golang.org/protobuf/proto"
)

type TestsGet struct {
	Tests   []*Test
	Success bool
}

func (a *TestsGet) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_GetTests,
	}
	return proto.Marshal(req)
}

func (a *TestsGet) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}

	a.Tests = resp.GetTests().Tests
	a.Success = resp.Success
	return nil
}
