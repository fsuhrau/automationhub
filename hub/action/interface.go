package action

type Interface interface {
	GetActionType() ActionType
	Serialize() ([]byte, error)
	ProcessResponse(response *Response) error
}
