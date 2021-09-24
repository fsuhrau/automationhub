package sse

type Event struct {
	Channel string
	Content interface{}
}