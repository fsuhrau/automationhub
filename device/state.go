package device

type State uint8
const (
	StateUnknown State = iota
	StateShutdown
	StateRemoteDisconnected
	StateBooted
	StateLocked
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
		return "remote disconnected"
	case StateLocked:
		return "locked"
	}
	return ""
}

type ConnectionState uint8
const (
	Disconnected ConnectionState = iota
	Connected
)

