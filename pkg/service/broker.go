package service

import (
	"errors"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

var (
	ErrNoBrokerConfigured = errors.New("broker: no broker configured")
	ErrIllegalUnsubscribe = errors.New("broker: unsubscribe without prior subscription illegal")
)

// Context is the structure of the data that is passed to a channel.
type Context struct {
	Service    *Service
	Cloudevent *cloudevents.Event
}

// Channel contains basic information about a service channel.
type Channel struct {
	Name        string `json:"name"`
	Subscribers int    `json:"subscribers"`
}

// ChannelHandler describes the function signature of the functions
// that implements the logic for the channel.
type ChannelHandler func(*Context) error

// Broker is an abstraction to allow provider agnostic interactions
// with a event broker or message queue.
type Broker interface {
	Bind(*Service)

	Subscribe(string, ChannelHandler) error
	Unsubscribe(string) error
	Publish(string, interface{}) error

	Request(string, interface{}) (*Context, error)

	// TODO: Abstract message queue / stream API.

	Connect() error
	Disconnect() error
}
