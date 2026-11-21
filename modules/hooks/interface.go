package hooks

type Level int

const (
	LevelSuccess Level = iota
	LevelUnstable
	LevelError
)

// Message is what a hook sends: a title, a text, an optional link and a level.
type Message struct {
	Title string
	Text  string
	Link  string
	Level Level
}

type Hook interface {
	Send(msg Message)
}
