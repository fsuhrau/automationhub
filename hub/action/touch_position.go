package action

type TouchDownPosition struct {
	PosX    int64
	PosY    int64
	Success bool
}

func (a *TouchDownPosition) GetActionType() ActionType {
	return ActionType_TouchDown
}

func (a *TouchDownPosition) Serialize() ([]byte, error) {
	return serialize(ActionType_TouchDown, touchPayload(a.PosX, a.PosY))
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
	return serialize(ActionType_TouchUp, touchPayload(a.PosX, a.PosY))
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
	return serialize(ActionType_TouchMove, touchPayload(a.PosX, a.PosY))
}

func (a *TouchMovePosition) ProcessResponse(response *Response) error {
	a.Success = response.Success
	return nil
}
