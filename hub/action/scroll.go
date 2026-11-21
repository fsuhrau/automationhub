package action

import "github.com/fsuhrau/automationhub/device/appproto"

type Scroll struct {
	OffsetX int64
	OffsetY int64
}

func (a *Scroll) GetActionType() appproto.ActionType {
	return appproto.ActionType_Move
}

func (a *Scroll) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_Move, appproto.RequestData{MoveOffset: &appproto.MoveOffset{From: a.OffsetX, To: a.OffsetY}})
}

func (a *Scroll) ProcessResponse(response *appproto.Response) error {
	return nil
}
