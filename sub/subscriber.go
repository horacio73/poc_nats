package main

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/nats-io/nats.go"
)

// Sub é uma função que consome do stream
func Sub() {
	// Create an Event.
	event := cloudevents.NewEvent()
	event.MarshalJSON()

	// Connect to a server
	nc, err := nats.Connect(nats.DefaultURL)

	if err != nil {
		return
	}

	// Simple Async Subscriber
	nc.Subscribe("foo", func(m *nats.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	})

	// Responding to a request message
	nc.Subscribe("request", func(m *nats.Msg) {
		m.Respond([]byte("answer is 42"))
	})
}
