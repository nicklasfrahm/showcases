package service

import (
	"errors"
)

var (
	ErrIllegalUnsubscribe = errors.New("broker: illegal unsubscribe")
)

// Message is an abstract interface that provides access to
// to the specific message type of the broker implementation.
type Message struct {
	Endpoint *string
	Reply    *string
	Data     *[]byte

	// TODO: Benchmark this. Possible point of optimization.
	Service *Service
}

// MessageHandler describes the function signature of message
// handler. Note that we do not need a pointer here as the
// Message type is an interface.
type MessageHandler func(m *Message)

// Broker is an abstraction to allow provider agnostic interactions
// with a event broker or message queue.
type Broker interface {
	Bind(*Service)

	Subscribe(string, MessageHandler) error
	Unsubscribe(string) error
	Publish(string, []byte) error

	Request(string, []byte) (*Message, error)

	Connect() error
}
