package main

import (
	"encoding/json"
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
	svc.BrokerEndpoint(">", func(ctx *service.Context) {
		subject := ctx.Cloudevent.Type()

		// Omit response messages as they should be distributed via a dedicated type.
		if subject == "response" {
			return
		}

		// Park data in interface to allow encoding as JSON.
		var data interface{}
		if err := ctx.Cloudevent.DataAs(&data); err != nil {
			ctx.Service.Logger.Error().Err(err).Msgf("Failed to load data from cloud event")
			return
		}

		if data == nil {
			// There is no data. Just log the subject.
			ctx.Service.Logger.Info().Msgf("%s", subject)
			return
		}

		encoded, err := json.MarshalIndent(data, "", " ")
		if err != nil {
			ctx.Service.Logger.Error().Err(err).Msgf("Failed to encode data from cloud event")
		}

		ctx.Service.Logger.Info().Msgf("%s\n%s", subject, string(encoded))
	})

	// Wait until error occurs or signal is received.
	svc.Start()
}
