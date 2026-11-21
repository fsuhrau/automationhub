package device

type State uint8

const (
	StateNone State = iota
	StateUnknown
	StateShutdown
	// StateRemoteDisconnected: the node doesn't see the device, e.g. an
	// unplugged USB phone or a network device that isn't connected
	StateRemoteDisconnected
	StateBooted
	StateLocked
	StateNodeDisconnected
	// StateUnauthorized: connected, but USB debugging isn't accepted on the device yet
	StateUnauthorized
)

func StateToString(state State) string {
	switch state {
	case StateUnknown:
		return "unknown"
	case StateShutdown:
		return "shutdown"
	case StateBooted:
		return "booted"
	case StateRemoteDisconnected:
		return "disconnected"
	case StateLocked:
		return "locked"
	case StateNodeDisconnected:
		return "node disconnected"
	case StateUnauthorized:
		return "unauthorized"
	}
	return ""
}

// ParseState maps a state name, as the platforms store it and the nodes send
// it, to its State. Unknown names give StateUnknown. StateLocked and StateNone
// have no name on the wire.
//
// This is the union of the platform mappings that exist today (androiddevice,
// iosdevice, iossim, macos, web, node, unityeditor). They all map an unknown
// name to StateUnknown, except unityeditor: it maps it to StateShutdown (G27
// switches it to ParseState).
func ParseState(name string) State {
	switch name {
	case "StateShutdown":
		return StateShutdown
	case "StateRemoteDisconnected":
		return StateRemoteDisconnected
	case "StateBooted":
		return StateBooted
	case "StateNodeDisconnected":
		return StateNodeDisconnected
	case "StateUnauthorized":
		return StateUnauthorized
	}
	return StateUnknown
}

type ConnectionState uint8

const (
	Disconnected ConnectionState = iota
	Connected
)
