package action

import "encoding/json"

type TouchDownPosition struct {
	PosX    int64
	PosY    int64
	Success bool
}

func (a *TouchDownPosition) GetActionType() ActionType {
	return ActionType_TouchDown
}

func (a *TouchDownPosition) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_TouchDown,
		Payload:    RequestData{Touch: &Touch{Xoffset: a.PosX, Yoffset: a.PosY}},
	}
	return json.Marshal(req)
}

func (a *TouchDownPosition) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}

type TouchUpPosition struct {
	PosX    int64
	PosY    int64
	Success bool
}

func (a *TouchUpPosition) GetActionType() ActionType {
	return ActionType_TouchUp
}

func (a *TouchUpPosition) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_TouchUp,
		Payload:    RequestData{Touch: &Touch{Xoffset: a.PosX, Yoffset: a.PosY}},
	}
	return json.Marshal(req)
}

func (a *TouchUpPosition) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}

type TouchMovePosition struct {
	PosX    int64
	PosY    int64
	Success bool
}

func (a *TouchMovePosition) GetActionType() ActionType {
	return ActionType_TouchMove
}

func (a *TouchMovePosition) Serialize() ([]byte, error) {
	req := &Request{
		ActionType: ActionType_TouchMove,
		Payload:    RequestData{Touch: &Touch{Xoffset: a.PosX, Yoffset: a.PosY}},
	}
	return json.Marshal(req)
}

func (a *TouchMovePosition) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
