package sse

type Publisher interface {
	PublishEvent(event Event)
}