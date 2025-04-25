package events

var AppCreated appCreated

type AppCreatedPayload struct {
	AppID uint        `json:"appId"`
	App   interface{} `json:"app"`
}

type appCreated struct {
	handlers []interface{ Handle(AppCreatedPayload) }
}

func (u *appCreated) Register(handler interface{ Handle(AppCreatedPayload) }) {
	u.handlers = append(u.handlers, handler)
}

func (u appCreated) Trigger(payload AppCreatedPayload) {
	for _, handler := range u.handlers {
		go handler.Handle(payload)
	}
}
