package action

import "github.com/fsuhrau/automationhub/device/appproto"

type IsDisplayed struct {
	ElementID   string
	Success     bool
	IsDisplayed bool
}

func (a *IsDisplayed) GetActionType() appproto.ActionType {
	return appproto.ActionType_ElementIsDisplayed
}

func (a *IsDisplayed) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_ElementIsDisplayed, appproto.RequestData{Id: &a.ElementID})
}

func (a *IsDisplayed) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	visible, err := payloadValue(response.Payload.Visible, "visible")
	if err != nil {
		return err
	}
	a.IsDisplayed = visible
	return nil
}
