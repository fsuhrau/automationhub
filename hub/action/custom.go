package action

import (
	"google.golang.org/protobuf/proto"
)

type Custom struct {
	RequestData string
	Content []byte
	Success bool
}

func (a *Custom) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_Custom,
		Payload: &Request_Data{Data: a.RequestData},
	}
	return proto.Marshal(req)
}

func (a *Custom) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Content = resp.GetData()
	a.Success = resp.Success
	return nil
}
