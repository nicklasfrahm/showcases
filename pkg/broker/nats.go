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
	if broker.natsConn == nil || broker.service == nil {
		broker.queuedSubscriptions[endpoint] = endpointHandler
		return nil
	}

	subscription, err := broker.natsConn.QueueSubscribe(endpoint, broker.service.Config.Name, func(msg *nats.Msg) {
		event := cloudevents.NewEvent()
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			broker.service.Logger.Error().Err(err).Msg("Failed to decode cloud event")
			return
		}

		// Rewrite source to response topic.
		event.SetSource(msg.Reply)

		// Invoke the endpoint handler with the user-defined business logic.
		endpointHandler(&service.Context{
			Service:    broker.service,
			Cloudevent: &event,
		})
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

	// Configure reconnection handlers.
	extraOptions := []nats.Option{
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			broker.service.Logger.Warn().Err(err).Msgf("Disconnected from broker: %s", redactedBrokerURI)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			broker.service.Logger.Info().Msgf("Reconnected to broker: %s", redactedBrokerURI)
		}),
	}
	broker.options.NATSOptions = append(broker.options.NATSOptions, extraOptions...)

	// Connect to NATS broker.
	broker.service.Logger.Info().Msgf("Broker URI: %s", redactedBrokerURI)
	natsConn, err := nats.Connect(broker.options.URI, broker.options.NATSOptions...)
	if err != nil {
		return err
	}
	broker.natsConn = natsConn

	// Create JetStream context.
	jetstreamCtx, err := natsConn.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return err
	}
	broker.jetstreamCtx = jetstreamCtx

	// Subscribe to queued subscriptions.
	for endpoint := range broker.queuedSubscriptions {
		if err := broker.Subscribe(endpoint, broker.queuedSubscriptions[endpoint]); err != nil {
			return err
		}

		delete(broker.queuedSubscriptions, endpoint)
	}

	return nil
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
