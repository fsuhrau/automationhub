package action

import (
	"google.golang.org/protobuf/proto"
)

type IsDisplayed struct {
	ElementID string
	Success bool
	IsDisplayed bool
}

func (a *IsDisplayed) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_ElementIsDisplayed,
		Payload: &Request_Id{Id: a.ElementID},
	}
	return proto.Marshal(req)
}

func (a *IsDisplayed) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Success = resp.Success
	a.IsDisplayed = resp.GetVisible()
	return nil
}

