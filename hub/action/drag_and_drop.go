package action

import "github.com/fsuhrau/automationhub/device/appproto"

type DragAndDrop struct {
	From string
	To   string
}

func (a *DragAndDrop) GetActionType() appproto.ActionType {
	return appproto.ActionType_DragAndDrop
}

func (a *DragAndDrop) Serialize() ([]byte, error) {
	return serialize(appproto.ActionType_DragAndDrop, appproto.RequestData{MoveElement: &appproto.MoveElement{From: a.From, To: a.To}})
}

func (a *DragAndDrop) ProcessResponse(response *appproto.Response) error {
	return nil
}
