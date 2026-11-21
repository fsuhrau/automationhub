package action

import "github.com/fsuhrau/automationhub/device/appproto"

type UnityReset struct {
	Assembly string
	Class    string
	Method   string
	Success  bool
}

func (a *UnityReset) GetActionType() appproto.ActionType {
	return appproto.ActionType_UnityReset
}

func (a *UnityReset) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_UnityReset, appproto.RequestData{})
}

func (a *UnityReset) ProcessResponse(response *appproto.Response) error {
	a.Success = response.Success
	return nil
}
