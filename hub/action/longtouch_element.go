package action

import "github.com/fsuhrau/automationhub/device/appproto"

type LongTouchElement struct {
	ElementID string
	Success   bool
}

func (a *LongTouchElement) GetActionType() appproto.ActionType {
	return appproto.ActionType_LongTouch
}

func (a *LongTouchElement) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_LongTouch, appproto.RequestData{Id: &a.ElementID})
}

func (a *LongTouchElement) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	return nil
}
