package device

type State uint8

const (
	StateNone State = iota
	StateUnknown
	StateShutdown
	StateRemoteDisconnected
	StateBooted
	StateLocked
	StateNodeDisconnected
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
	case StateNodeDisconnected:
		return "node disconnected"
	}
	return ""
}

type ConnectionState uint8

const (
	Disconnected ConnectionState = iota
	Connected
)
