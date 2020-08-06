package action

import (
	"google.golang.org/protobuf/proto"
)

type TouchDownPosition struct {
	PosX int64
	PosY int64
	Success bool
}

func (a *TouchDownPosition) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_TouchDown,
		Payload: &Request_Touch{&Touch{Xoffset: a.PosX, Yoffset: a.PosY}},
	}
	return proto.Marshal(req)
}

func (a *TouchDownPosition) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Success = resp.Success
	return nil
}

type TouchUpPosition struct {
	PosX int64
	PosY int64
	Success bool
}

func (a *TouchUpPosition) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_TouchUp,
		Payload: &Request_Touch{&Touch{Xoffset: a.PosX, Yoffset: a.PosY}},
	}
	return proto.Marshal(req)
}

func (a *TouchUpPosition) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Success = resp.Success
	return nil
}


type TouchMovePosition struct {
	PosX int64
	PosY int64
	Success bool
}

func (a *TouchMovePosition) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_TouchMove,
		Payload: &Request_Touch{&Touch{Xoffset: a.PosX, Yoffset: a.PosY}},
	}
	return proto.Marshal(req)
}

func (a *TouchMovePosition) Deserialize(content []byte) error {
	resp := &Response{}
	if err := proto.Unmarshal(content, resp); err != nil {
		return err
	}
	a.Success = resp.Success
	return nil
}