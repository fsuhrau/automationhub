package device

type State uint8

const (
	Unknown State = iota
	Shutdown
	RemoteDisconnected
	Booted
)

func StateToString(state State) string {
	switch state {
	case Unknown:
		return "unknown"
	case Shutdown:
		return "shutdown"
	case Booted:
		return "booted"
	case RemoteDisconnected:
		return "remote disconnected"
	}
	return ""
}

type ConnectionState uint8

const (
	Disconnected ConnectionState = iota
	Connected
)

type ApplicationType uint8

const (
	iOS ApplicationType = iota
	iOSSim
	MacOS
	Android
)
