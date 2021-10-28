package broker

import (
	"encoding/json"
	"net/url"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/nicklasfrahm/showcases/pkg/service"
)

// This file contains the implementation of the Broker interface
// for the NATS event broker and message queue (https://nats.io/).

type NATSOptions struct {
	URI            string
	NATSOptions    []nats.Option
	RequestTimeout time.Duration
}

type NATS struct {
	service             *service.Service
	options             *NATSOptions
	natsConn            *nats.Conn
	jetstreamCtx        nats.JetStreamContext
	activeSubscriptions map[string]*nats.Subscription
	queuedSubscriptions map[string]service.EndpointHandler
	mutex               *sync.Mutex
}

func NewNATS(opts *NATSOptions) service.Broker {
	// TODO: Create a PreEndpoint and PostEndpoint method for better readability.

	if opts.RequestTimeout == 0 {
		opts.RequestTimeout = 1000 * time.Millisecond
	}

	return &NATS{
		options:             opts,
		activeSubscriptions: make(map[string]*nats.Subscription),
		queuedSubscriptions: make(map[string]service.EndpointHandler),
		mutex:               &sync.Mutex{},
	}
}

func (broker *NATS) Bind(svc *service.Service) {
	broker.service = svc
}

func (broker *NATS) Subscribe(endpoint string, endpointHandler service.EndpointHandler) error {
	// Queue subscriptions that are made before connecting to the server.
	if broker.natsConn == nil {
		broker.queuedSubscriptions[endpoint] = endpointHandler
		return nil
	}

	subscription, err := broker.natsConn.Subscribe(endpoint, func(msg *nats.Msg) {
		// Invoke the endpoint handler with the user-defined business logic.
		endpointHandler(broker.newContext(msg))
	})
	if err != nil {
		return err
	}

	broker.mutex.Lock()
	broker.activeSubscriptions[endpoint] = subscription
	broker.mutex.Unlock()

	return nil
}

func (b *NATS) Unsubscribe(endpoint string) error {
	// Notify developer that there is a logic error
	// when unsubscribing without prior subscription.
	if b.activeSubscriptions[endpoint] == nil {
		return service.ErrIllegalUnsubscribe
	}

	return b.activeSubscriptions[endpoint].Unsubscribe()
}

func (broker *NATS) Publish(endpoint string, event interface{}) error {
	encoded, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return broker.natsConn.Publish(endpoint, encoded)
}

func (broker *NATS) Request(endpoint string, event interface{}) (*service.Context, error) {
	encoded, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	msg, err := broker.natsConn.Request(endpoint, encoded, broker.options.RequestTimeout)
	if err != nil {
		return nil, err
	}

	return broker.newContext(msg), nil
}

func (b *NATS) Connect() error {
	// Ensure that URI is provided.
	if b.options.URI == "" {
		b.service.Logger.Fatal().Msgf("Configuration missing: BROKER_URI")
	}

	// Parse URI to redact secrets.
	redacted, err := url.Parse(b.options.URI)
	if err != nil {
		b.service.Logger.Fatal().Msgf("Configuration invalid: BROKER_URI")
	}
	// Manually redact username and password rather than replacing it with xxx.
	redacted.User = &url.Userinfo{}

	// Connect to NATS broker.
	b.service.Logger.Info().Msg("Connecting to: " + redacted.String())
	nc, err := nats.Connect(b.options.URI, b.options.NATSOptions...)
	if err != nil {
		return err
	}
	b.natsConn = nc

	// Create JetStream context.
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return err
	}
	b.jetstreamCtx = js

	// Subscribe to queued subscriptions.
	for endpoint := range b.queuedSubscriptions {
		if err := b.Subscribe(endpoint, b.queuedSubscriptions[endpoint]); err != nil {
			return err
		}

		delete(b.queuedSubscriptions, endpoint)
	}

	return nil
}

func (broker *NATS) newContext(msg *nats.Msg) *service.Context {
	// Create cloud event for NATS message.
	event := broker.service.NewEvent()
	event.SetSource(msg.Reply)
	event.SetType(msg.Subject)
	event.SetData(msg.Data)

	return &service.Context{
		Service:    broker.service,
		Cloudevent: event,
	}
}
