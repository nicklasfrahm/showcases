package main

import (
	"os"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/nicklasfrahm/showcases/pkg/broker"
	"github.com/nicklasfrahm/showcases/pkg/service"
)

var (
	name    = "unknown"
	version = "dev"
)

func main() {
	// Create new service instance.
	svc := service.New(service.Config{
		Name:    name,
		Version: version,
	})

	// Configure broker connection.
	svc.UseBroker(broker.NewNATS(&broker.NATSOptions{
		URI: os.Getenv("BROKER_URI"),
		NATSOptions: []nats.Option{
			nats.Name(name),
			nats.Timeout(1 * time.Second),
			nats.PingInterval(5 * time.Second),
			nats.MaxPingsOutstanding(6),
		},
		RequestTimeout: 20 * time.Millisecond,
	}))

	svc.BrokerEndpoint("v1.mail.create", func(m *service.Message) {
		// TODO: Implement actual mail sending logic.
		res := []byte("test")

		if err := m.Service.Broker.Publish(*m.Reply, res); err != nil {
			m.Service.Logger.Error().Err(err).Msgf("Failed to reply: %s", *m.Endpoint)
		}
	})

	// Wait until error occurs or signal is received.
	svc.Start()
}
