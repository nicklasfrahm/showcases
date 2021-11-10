package main

import (
	"encoding/json"
	"os"
	"regexp"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/nicklasfrahm/showcases/pkg/broker"
	"github.com/nicklasfrahm/showcases/pkg/service"
)

var (
	name    = "unknown"
	version = "dev"
)

var (
	channelsRegExp = regexp.MustCompile(`^channels\..*`)
)

func main() {
	// Create new service instance.
	svc := service.New(service.Config{
		Name:    name,
		Version: version,
	})

	// Configure broker connection.
	svc.UseBroker(broker.NewNATS(&broker.NATSOptions{
		URI:            os.Getenv("BROKER_URI"),
		RequestTimeout: 20 * time.Millisecond,
	}))

	// Define catch-all channel for audit service.
	svc.BrokerChannel(">", CatchAll())

	// Wait until error occurs or signal is received.
	svc.Start()
}

func CatchAll() service.ChannelHandler {
	return func(ctx *service.Context) error {
		channel := ctx.Cloudevent.Type()

		// Omit logging the content for the following channels:
		// - response
		// - channels.*
		if channel == "response" || channelsRegExp.MatchString(channel) {
			return nil
		}

		// Park data in interface to allow encoding as JSON.
		var data interface{}
		if err := ctx.Cloudevent.DataAs(&data); err != nil {
			return err
		}

		if data == nil {
			// There is no data. Just log the subject.
			ctx.Service.Logger.Info().Msgf("%s", channel)
			return nil
		}

		encoded, err := json.MarshalIndent(data, "", " ")
		if err != nil {
			return err
		}

		ctx.Service.Logger.Info().Msgf("%s\n%s", channel, string(encoded))
		return nil
	}
}
