package broker

import (
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
	queuedSubscriptions map[string]service.MessageHandler
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
		queuedSubscriptions: make(map[string]service.MessageHandler),
		mutex:               &sync.Mutex{},
	}
}

func (b *NATS) Bind(svc *service.Service) {
	b.service = svc
}

func (b *NATS) Subscribe(endpoint string, messageHandler service.MessageHandler) error {
	// Queue subscriptions that are made before connecting to the server.
	if b.natsConn == nil {
		b.queuedSubscriptions[endpoint] = messageHandler
		return nil
	}

	subscription, err := b.natsConn.Subscribe(endpoint, func(msg *nats.Msg) {
		messageHandler(&service.Message{
			Endpoint: &msg.Subject,
			Reply:    &msg.Reply,
			Data:     &msg.Data,
			Service:  b.service,
		})
	})
	if err != nil {
		return err
	}

	b.mutex.Lock()
	b.activeSubscriptions[endpoint] = subscription
	b.mutex.Unlock()

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

func (b *NATS) Publish(endpoint string, data []byte) error {
	err := b.natsConn.Publish(endpoint, data)
	return err
}

func (b *NATS) Request(endpoint string, data []byte) (*service.Message, error) {
	msg, err := b.natsConn.Request(endpoint, data, b.options.RequestTimeout)
	if err != nil {
		return nil, err
	}

	return &service.Message{
		Endpoint: &msg.Subject,
		Reply:    &msg.Reply,
		Data:     &msg.Data,
		Service:  b.service,
	}, nil
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
