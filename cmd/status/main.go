package main

import (
	"os"
	"time"

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
		URI:            os.Getenv("BROKER_URI"),
		RequestTimeout: 20 * time.Millisecond,
	}))

	svc.BrokerChannel("channels.create", ChannelsCreate)
	svc.BrokerChannel("channels.find", ChannelsFind)
	svc.BrokerChannel("channels.delete", ChannelsDelete)

	// Wait until error occurs or signal is received.
	svc.Start()
}
