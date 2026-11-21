package action

import "github.com/fsuhrau/automationhub/device/appproto"

type TouchElement struct {
	ElementID string
	Success   bool
}

func (a *TouchElement) GetActionType() appproto.ActionType {
	return appproto.ActionType_ElementTouch
}

func (a *TouchElement) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_ElementTouch, appproto.RequestData{Id: &a.ElementID})
}

func (a *TouchElement) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	return nil
}
