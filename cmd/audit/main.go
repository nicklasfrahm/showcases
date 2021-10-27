package main

import (
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
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

	// Define catch-all endpoint for audit service.
	svc.BrokerEndpoint(">", func(m *service.Message) {
		m.Service.Logger.Info().Msgf("%s >> %s", *m.Endpoint, string(*m.Data))
	})

	// Wait until error occurs or signal is received.
	svc.Start()
}
