package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

var (
	ErrNoGatewayConfigured = errors.New("gateway: no gateway configured")
)

// Request is an abstract interface that provides access to
// to high level properties of the broker implementation.
type Request struct {
	Ctx *fiber.Ctx

	Service *Service
}

// RequestHandler is the abstracted method that will be implemented
// by the API consumer.
type RequestHandler func(*Request) error

// TODO: Create abstraction such that implementation just
// uses enriched TCP, such that a gateway could also support
// MQTT or other protocols in the future.
type Gateway interface {
	Bind(*Service)

	Route(string, RequestHandler)

	Listen()
}
