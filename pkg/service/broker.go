package service

import (
	"errors"
)

var (
	ErrNoBrokerConfigured = errors.New("broker: no broker configured")
	ErrIllegalUnsubscribe = errors.New("broker: unsubscribe without prior subscription illegal")
)

// Broker is an abstraction to allow provider agnostic interactions
// with a event broker or message queue.
type Broker interface {
	Bind(*Service)

	Subscribe(string, EndpointHandler) error
	Unsubscribe(string) error
	Publish(string, interface{}) error

	Request(string, interface{}) (*Context, error)

	// TODO: Abstract message queue / stream API.

	Connect() error
}
