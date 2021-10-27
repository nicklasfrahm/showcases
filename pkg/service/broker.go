package service

import (
	"errors"
)

var (
	ErrNoBrokerConfigured = errors.New("broker: no broker configured")
	ErrIllegalUnsubscribe = errors.New("broker: unsubscribe without prior subscription illegal")
)

// Request is an abstract struct that provides access to
// to high level properties of the broker implementation.
type Message struct {
	Endpoint *string
	Reply    *string
	Data     *[]byte

	// TODO: Benchmark this. Possible point of optimization.
	Service *Service
}

// MessageHandler describes the function signature of message
// handler.
type MessageHandler func(*Message)

// Broker is an abstraction to allow provider agnostic interactions
// with a event broker or message queue.
type Broker interface {
	Bind(*Service)

	Subscribe(string, MessageHandler) error
	Unsubscribe(string) error
	Publish(string, []byte) error

	Request(string, []byte) (*Message, error)

	// TODO: Abstract message queue / stream API.

	Connect() error
}
