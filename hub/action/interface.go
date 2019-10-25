package action

type Interface interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}
