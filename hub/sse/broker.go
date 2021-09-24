package sse

import (
	"github.com/gin-gonic/gin"
	"log"
)

type Broker struct {
	// Events are pushed to this channel by the main events-gathering routine
	Events chan Event

	// New client connections
	NewClients chan chan Event

	// Closed client connections
	ClosedClients chan chan Event

	// Total client connections
	TotalClients map[chan Event]bool
}

// New event messages are broadcast to all registered client connection channels
type ClientChan chan Event

// Initialize event and Start procnteessing requests
func NewBroker() (broker *Broker) {
	broker = &Broker{
		Events:        make(chan Event),
		NewClients:    make(chan chan Event),
		ClosedClients: make(chan chan Event),
		TotalClients:  make(map[chan Event]bool),
	}

	go broker.listen()
	return
}

//It Listens all incoming requests from clients.
//Handles addition and removal of clients and broadcast messages to clients.
func (stream *Broker) listen() {
	for {
		select {
		// Add new available client
		case client := <-stream.NewClients:
			stream.TotalClients[client] = true
			log.Printf("Client added. %d registered clients", len(stream.TotalClients))

		// Remove closed client
		case client := <-stream.ClosedClients:
			delete(stream.TotalClients, client)
			log.Printf("Removed client. %d registered clients", len(stream.TotalClients))

		// Broadcast message to client
		case eventMsg := <-stream.Events:
			log.Printf("Send Event to %d clients", len(stream.TotalClients))
			for clientMessageChan := range stream.TotalClients {
				clientMessageChan <- eventMsg
			}
		}
	}
}

func (stream *Broker) ServeHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		clientChan := make(ClientChan)

		// Send new connection to event server
		stream.NewClients <- clientChan

		defer func() {
			// Send closed connection to event server
			stream.ClosedClients <- clientChan
		}()

		go func() {
			// Send connection that is closed by client to event server
			<-c.Done()
			stream.ClosedClients <- clientChan
		}()

		c.Next()
	}
}