package main

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nats-io/nats.go"

	"github.com/nicklasfrahm/virtual-white-board/pkg/service"
)

var (
	name    = "unknown"
	version = "dev"
)

func main() {
	natsUri := os.Getenv("NATS_URI")

	// Create new service instance.
	svc := service.New(service.Config{
		Name:    name,
		Version: version,
	})

	// Configure broker connection.
	svc.UseBroker(natsUri)

	// Subscribe to wildcard.
	svc.Broker.Subscribe(">", func(m *nats.Msg) {
		svc.Logger.Info().Msgf("%s >> %s", m.Subject, string(m.Data))
	})

	// Wait until error occurs or signal is received.
	svc.Listen()
}
