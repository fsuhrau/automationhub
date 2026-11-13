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

type ConnectionState uint8

const (
	Disconnected ConnectionState = iota
	Connected
)
