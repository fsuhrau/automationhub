package action

import "github.com/fsuhrau/automationhub/device/appproto"

type SetValue struct {
	ElementID string
	Success   bool
	Attr      string
	Value     string
}

func (a *SetValue) GetActionType() appproto.ActionType {
	return appproto.ActionType_ElementSetValue
}

func (a *SetValue) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_ElementSetValue, appproto.RequestData{SetAttr: &appproto.SetAttr{Id: a.ElementID, Attr: a.Attr, Val: a.Value}})
}

func (a *SetValue) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	return nil
}
