package events

var AppCreated = NewBus[AppCreatedPayload]() //nolint:gochecknoglobals // event bus registry (concept C4)

type AppCreatedPayload struct {
	AppID uint        `json:"appId"`
	App   interface{} `json:"app"`
}
