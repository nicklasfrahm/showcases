package service

// Request is an abstract interface that provides access to
// to high level properties of the broker implementation.

// Gateway is an abstraction to allow provider agnostic interactions
// with a TCP server.
type Gateway interface {
	Bind(*Service)

	Route(string) error

	Listen()
}
