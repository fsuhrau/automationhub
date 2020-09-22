package action

import (
	"google.golang.org/protobuf/proto"
)

type DragAndDrop struct {
	From string
	To string
}

func (a *DragAndDrop) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_DragAndDrop,
		Payload: &Request_MoveElement{&MoveElement{From: a.From, To:a.To}},
	}
	return proto.Marshal(req)
}

func (a *DragAndDrop) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	//if !resp.Success {
	//	return fmt.Errorf("drag and drop failed")
	//}
	return nil
}
