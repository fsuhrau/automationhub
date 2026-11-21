package action

import "github.com/fsuhrau/automationhub/device/appproto"

type Interface interface {
	GetActionType() appproto.ActionType
	Serialize() ([]byte, error)
	ProcessResponse(response *appproto.Response) error
}
