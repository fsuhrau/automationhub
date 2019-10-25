package devices

type State uint8

const (
	Unknown State = iota
	Shutdown
	Booted
)

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
