package broker

import (
	"encoding/json"
	"net/url"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/nicklasfrahm/showcases/pkg/service"
)

// This file contains the implementation of the Broker interface
// for the NATS event broker and message queue (https://nats.io/).

// TODO: Add a `.Reply()` function to the broker that handles replying more elegantly.
// TODO: Add a `.Broadcast()` function that enables or disables broadcasting after handler completion.
// TODO: Create a canonical channel format for responses (`*.response`), handler success (`*.success`) and handler failure (`*.failure`).

const (
	ChannelSubscribe   = "channels.create"
	ChannelUnsubscribe = "channels.delete"
)

type NATSOptions struct {
	URI            string
	NATSOptions    []nats.Option
	RequestTimeout time.Duration
}

type NATS struct {
	service             *service.Service
	options             *NATSOptions
	natsConn            *nats.Conn
	activeSubscriptions map[string]*nats.Subscription
	queuedSubscriptions map[string]service.ChannelHandler
	mutex               *sync.Mutex
}

func NewNATS(opts *NATSOptions) service.Broker {
	if opts.RequestTimeout == 0 {
		opts.RequestTimeout = 1000 * time.Millisecond
	}

	return &NATS{
		options:             opts,
		activeSubscriptions: make(map[string]*nats.Subscription),
		queuedSubscriptions: make(map[string]service.ChannelHandler),
		mutex:               &sync.Mutex{},
	}
}

func (broker *NATS) Bind(svc *service.Service) {
	broker.service = svc
}

func (broker *NATS) Subscribe(channel string, channelHandler service.ChannelHandler) error {
	// Queue subscriptions that are made before connecting to the server.
	if broker.natsConn == nil || broker.service == nil {
		broker.queuedSubscriptions[channel] = channelHandler
		return nil
	}

	subscription, err := broker.natsConn.QueueSubscribe(channel, broker.service.Config.Name, func(msg *nats.Msg) {
		event := cloudevents.NewEvent()
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			broker.service.Logger.Error().Err(err).Msg("Failed to decode cloud event")
			return
		}

		// Rewrite source to response topic.
		event.SetSource(msg.Reply)

		// Invoke the channel handler with the user-defined business logic.
		if handlerErr := channelHandler(&service.Context{
			Service:    broker.service,
			Cloudevent: &event,
		}); handlerErr != nil {
			broker.service.Logger.Error().Err(handlerErr).Msg("Failed to run channel handler")
			return
		}
	})
	if err != nil {
		return err
	}

	broker.mutex.Lock()
	broker.activeSubscriptions[channel] = subscription
	broker.mutex.Unlock()

	// Attempt to register subscription.
	channelInfo := service.Channel{Name: channel}
	maxAttempts := 10
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		if _, err := broker.Request(ChannelSubscribe, channelInfo); err == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * time.Second)
	}
	// TODO: How to handle or log that there is no status service?

	return nil
}

func (broker *NATS) Unsubscribe(channel string) error {
	// Notify developer that there is a logic error
	// when unsubscribing without prior subscription.
	if broker.activeSubscriptions[channel] == nil {
		return service.ErrIllegalUnsubscribe
	}

	if err := broker.activeSubscriptions[channel].Unsubscribe(); err != nil {
		return err
	}

	// Attempt to register subscription.
	channelInfo := service.Channel{Name: channel}
	maxAttempts := 10
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		if _, err := broker.Request(ChannelUnsubscribe, channelInfo); err == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
	}
	// TODO: How to handle or log that there is no status service?

	return nil
}

func (broker *NATS) Publish(endpoint string, data interface{}) error {
	encoded, err := json.Marshal(broker.newEvent(endpoint, data))
	if err != nil {
		return err
	}

	return broker.natsConn.Publish(endpoint, encoded)
}

func (broker *NATS) Request(endpoint string, data interface{}) (*service.Context, error) {
	encoded, err := json.Marshal(broker.newEvent(endpoint, data))
	if err != nil {
		return nil, err
	}

	msg, err := broker.natsConn.Request(endpoint, encoded, broker.options.RequestTimeout)
	if err != nil {
		return nil, err
	}

	res := new(cloudevents.Event)
	if err := json.Unmarshal(msg.Data, res); err != nil {
		return nil, err
	}

	return &service.Context{
		Service:    broker.service,
		Cloudevent: res,
	}, nil
}

func (broker *NATS) Connect() error {
	// Ensure that URI is provided.
	if broker.options.URI == "" {
		broker.service.Logger.Fatal().Msgf("Configuration missing: BROKER_URI")
	}

	// Parse URI to redact secrets.
	redacted, err := url.Parse(broker.options.URI)
	if err != nil {
		broker.service.Logger.Fatal().Msgf("Configuration invalid: BROKER_URI")
	}
	// Manually redact username and password rather than replacing it with xxx.
	redacted.User = nil
	redactedBrokerURI := redacted.String()

	// Configure default options.
	defaultOptions := []nats.Option{
		nats.Name(broker.service.Config.Name),
		nats.Timeout(1 * time.Second),
		nats.PingInterval(5 * time.Second),
		nats.MaxPingsOutstanding(6),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			broker.service.Logger.Warn().Err(err).Msgf("Disconnected from broker: %s", redactedBrokerURI)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			broker.service.Logger.Info().Msgf("Reconnected to broker: %s", redactedBrokerURI)
		}),
	}
	broker.options.NATSOptions = append(defaultOptions, broker.options.NATSOptions...)

	// Connect to NATS broker.
	natsConn, err := nats.Connect(broker.options.URI, broker.options.NATSOptions...)
	if err != nil {
		return err
	}
	broker.natsConn = natsConn

	// Subscribe to queued subscriptions.
	for channel := range broker.queuedSubscriptions {
		if err := broker.Subscribe(channel, broker.queuedSubscriptions[channel]); err != nil {
			return err
		}

		delete(broker.queuedSubscriptions, channel)
	}

	// Log successful connection.
	broker.service.Logger.Info().Msgf("Connected to broker: %s", redactedBrokerURI)

	return nil
}

func (broker *NATS) Disconnect() error {
	// Close all subscriptions manually to ensure that the channels are unregistered.
	for channel := range broker.activeSubscriptions {
		if err := broker.Unsubscribe(channel); err != nil {
			return err
		}
	}

	// Drain connection.
	return broker.natsConn.Drain()
}

// newEvent is a convenience function that creates a new service-specific cloud event.
func (broker *NATS) newEvent(endpoint string, data interface{}) *cloudevents.Event {
	// Assemble new cloud event.
	event := cloudevents.NewEvent()
	event.SetID(uuid.NewString())
	event.SetSource(broker.service.Config.Name)
	event.SetData(cloudevents.ApplicationJSON, data)

	// Check if the event is directed towards a specific inbox.
	if strings.Split(endpoint, ".")[0] == "_INBOX" {
		event.SetType("response")
	} else {
		event.SetType(endpoint)
	}

	return &event
}
