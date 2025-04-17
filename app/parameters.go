package app

type AndroidParams struct {
	LaunchActivity string
}

type ExecutableParams struct {
	Executable string
}

type AppParams struct {
	AppBinaryID uint
	AppPath     string
	Additional  string
	Hash        string
	Size        int
	Android     *AndroidParams
	Executable  *ExecutableParams
}

type WebParams struct {
	StartURL string
}

type Parameter struct {
	Platform   string
	Identifier string
	Name       string
	Version    string
	App        *AppParams
	Web        *WebParams
}
