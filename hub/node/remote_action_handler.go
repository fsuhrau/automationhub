package node

import "github.com/fsuhrau/automationhub/device/remote"

// RemoteActionHandler lives in device/remote; the alias keeps the old name.
type RemoteActionHandler = remote.ActionHandler

// NewRemoteActionHandler is kept for the platform handlers.
var NewRemoteActionHandler = remote.NewActionHandler
