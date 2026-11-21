package action

import "github.com/fsuhrau/automationhub/device/appproto"

type Custom struct {
	RequestData string
	Content     []byte
	Success     bool
}

func (a *Custom) GetActionType() appproto.ActionType {
	return appproto.ActionType_Custom
}

func (a *Custom) Serialize() ([]byte, error) {
	data := []byte(a.RequestData)
	return serialize(appproto.ActionType_Custom, appproto.RequestData{Data: &data})
}

func (a *Custom) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	content, err := payloadValue(response.Payload.Data, "data")
	if err != nil {
		return err
	}
	a.Content = content
	return nil
}
