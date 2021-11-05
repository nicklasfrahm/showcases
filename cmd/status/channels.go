package main

import (
	"sync"

	"github.com/nicklasfrahm/showcases/pkg/service"
)

var channels = map[string]*service.Channel{}
var mutex = sync.Mutex{}

func ChannelsFind(ctx *service.Context) error {
	// Ensure safe concurrent access.
	mutex.Lock()

	// Create a list of a all channels.
	channelList := make([]service.Channel, len(channels))
	i := 0
	for _, channel := range channels {
		channelList[i] = *channel
		i += 1
	}

	// Release lock.
	mutex.Unlock()

	// Send reply. Please note that the source is an opaque string
	// that is used by the broker implementation to perform routing.
	if err := ctx.Service.Broker.Publish(ctx.Cloudevent.Source(), channelList); err != nil {
		return err
	}
	// Broadcast event.
	return ctx.Service.Broker.Publish("channels.found", channelList)
}

func ChannelsCreate(ctx *service.Context) error {
	// Decode event payload.
	channel := new(service.Channel)
	if err := ctx.Cloudevent.DataAs(channel); err != nil {
		return err
	}

	// Ensure safe concurrent access.
	mutex.Lock()

	// Check if the channel already exists.
	if channels[channel.Name] == nil {
		channels[channel.Name] = channel
	}
	channels[channel.Name].Subscribers += 1

	// Release lock.
	mutex.Unlock()

	// Send reply. Please note that the source is an opaque string
	// that is used by the broker implementation to perform routing.
	if err := ctx.Service.Broker.Publish(ctx.Cloudevent.Source(), channel); err != nil {
		return err
	}
	// Broadcast event.
	return ctx.Service.Broker.Publish("channels.created", channel)
}

func ChannelsDelete(ctx *service.Context) error {
	// Decode event payload.
	channel := new(service.Channel)
	if err := ctx.Cloudevent.DataAs(channel); err != nil {
		return err
	}

	// Ensure safe concurrent access.
	mutex.Lock()

	// Check if the channel exists.
	if channels[channel.Name] != nil {
		channels[channel.Name].Subscribers -= 1

		// Delete the channel if there are no subscribers.
		if channels[channel.Name].Subscribers == 0 {
			delete(channels, channel.Name)
		}
	}

	// Release lock.
	mutex.Unlock()

	// Send reply. Please note that the source is an opaque string
	// that is used by the broker implementation to perform routing.
	if err := ctx.Service.Broker.Publish(ctx.Cloudevent.Source(), channel); err != nil {
		return err
	}
	// Broadcast event.
	return ctx.Service.Broker.Publish("channels.deleted", channel)
}
