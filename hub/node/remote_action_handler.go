package node

import "github.com/fsuhrau/automationhub/device/remote"

// RemoteActionHandler lives in device/remote; the alias keeps the old name.
type RemoteActionHandler = remote.ActionHandler

// NewRemoteActionHandler is kept for the platform handlers.
//
//nolint:revive // argument-limit: signature is fixed by the platform handlers that call it
func NewRemoteActionHandler(masterURL, nodeIdentifier, deviceID string, authToken *string) *RemoteActionHandler {
	return remote.NewActionHandler(masterURL, nodeIdentifier, deviceID, authToken)
}
