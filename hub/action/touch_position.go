package action

import "github.com/fsuhrau/automationhub/device/appproto"

type TouchDownPosition struct {
	PosX    int64
	PosY    int64
	Success bool
}

func (a *TouchDownPosition) GetActionType() appproto.ActionType {
	return appproto.ActionType_TouchDown
}

func (a *TouchDownPosition) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_TouchDown, touchPayload(a.PosX, a.PosY))
}

func (a *TouchDownPosition) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	return nil
}

type TouchUpPosition struct {
	PosX    int64
	PosY    int64
	Success bool
}

func (a *TouchUpPosition) GetActionType() appproto.ActionType {
	return appproto.ActionType_TouchUp
}

func (a *TouchUpPosition) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_TouchUp, touchPayload(a.PosX, a.PosY))
}

func (a *TouchUpPosition) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	return nil
}

type TouchMovePosition struct {
	PosX    int64
	PosY    int64
	Success bool
}

func (a *TouchMovePosition) GetActionType() appproto.ActionType {
	return appproto.ActionType_TouchMove
}

func (a *TouchMovePosition) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_TouchMove, touchPayload(a.PosX, a.PosY))
}

func (a *TouchMovePosition) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	return nil
}
