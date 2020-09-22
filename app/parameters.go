package app

type Parameter struct {
	Name           string
	AppPath        string
	Identifier     string
	Version        string
	LaunchActivity string
	Additional     string
	Hash           [20]byte
}
