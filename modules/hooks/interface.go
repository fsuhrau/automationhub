package hooks

type Level int
const (
	LevelSuccess Level = iota
	LevelUnstable
	LevelError
)

type Hook interface {
	Send(title, message, link string, level Level)
}