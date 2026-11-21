package appproto

type ActionHandler interface {
	OnActionResponse(interface{}, *Response)
}
