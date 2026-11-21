package action

import "github.com/fsuhrau/automationhub/device/appproto"

type GetValue struct {
	ElementID string
	Success   bool
	Attr      string
	Value     string
}

func (a *GetValue) GetActionType() appproto.ActionType {
	return appproto.ActionType_ElementGetValue
}

func (a *GetValue) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_ElementGetValue, appproto.RequestData{GetAttr: &appproto.GetAttr{Id: a.ElementID, Attr: a.Attr}})
}

func (a *GetValue) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	value, err := payloadValue(response.Payload.Value, "value")
	if err != nil {
		return err
	}
	a.Value = value
	return nil
}
