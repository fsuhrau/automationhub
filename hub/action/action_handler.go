package action

type ActionHandler interface {
	OnActionResponse(interface{}, *Response)
}