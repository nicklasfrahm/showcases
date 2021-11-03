package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

// Config contains the configuration of a microservice.
type Config struct {
	Name    string
	Version string
}

// Service contains the state and configuration of a microservice.
type Service struct {
	Logger  *zerolog.Logger
	Broker  Broker
	Gateway Gateway
	Config  Config

	signals   chan os.Signal
	terminate chan bool
}

// Context is the structure of the data that is passed to an endpoint.
type Context struct {
	Service    *Service
	Cloudevent *cloudevents.Event
}

// EndpointHandler describes the function signature of the functiona
// that implements the logic for the service endpoint.
type EndpointHandler func(*Context)

// New returns a new service for the given configuration.
func New(config Config) *Service {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	return &Service{
		Logger: &log.Logger,
		Config: config,

		signals:   make(chan os.Signal, 1),
		terminate: make(chan bool, 1),
	}
}

func (svc *Service) UseBroker(b Broker) *Service {
	// Store reference to broker.
	svc.Broker = b

	// Pass service pointer to broker.
	svc.Broker.Bind(svc)

	// Return the service pointer to allow method chaining.
	return svc
}

func (svc *Service) UseGateway(g Gateway) *Service {
	// Store reference to gateway.
	svc.Gateway = g

	// Pass service pointer to gateway.
	svc.Gateway.Bind(svc)

	// Return the service pointer to allow method chaining.
	return svc
}

func (svc *Service) BrokerEndpoint(endpoint string, endpointHandler EndpointHandler) *Service {
	// Ensure that a broker is configured when endpoints are defined.
	if svc.Broker == nil {
		svc.Logger.Fatal().Err(ErrNoBrokerConfigured).Msg("Failed to register broker endpoint")
	}

	// Subscribe to broker endpoint.
	if err := svc.Broker.Subscribe(endpoint, endpointHandler); err != nil {
		svc.Logger.Fatal().Err(err).Msgf("Failed to register broker endpoint")
	}

	// Log the registered endpoint.
	svc.Logger.Info().Msg("Endpoint registered: " + endpoint)

	// Return the service pointer to allow method chaining.
	return svc
}

func (svc *Service) GatewayEndpoint(requestHandler RequestHandler) *Service {
	// Ensure that a gateway is configured when endpoints are defined.
	if svc.Gateway == nil {
		svc.Logger.Fatal().Err(ErrNoGatewayConfigured).Msg("Failed to register gateway endpoint")
	}

	// Pass request handler to gateway.
	svc.Gateway.Route(requestHandler)

	// Return the service pointer to allow method chaining.
	return svc
}

// Start is a blocking function that starts the service.
func (svc *Service) Start() {
	// Subscribe to OS signals and asynchronously await them in goroutine.
	signal.Notify(svc.signals, syscall.SIGINT, syscall.SIGTERM)
	go svc.awaitSignals()

	// Log basic service information.
	svc.Logger.Info().Msgf("Service: %s", svc.Config.Name)
	svc.Logger.Info().Msgf("Version: %s", svc.Config.Version)

	// Connect to broker if configured.
	if svc.Broker != nil {
		if err := svc.Broker.Connect(); err != nil {
			svc.Logger.Fatal().Err(err).Msg("Failed to connect to broker")
		}
	}

	// Run blocking gateway in goroutine.
	if svc.Gateway != nil {
		go svc.Gateway.Listen()
	}

	// Block until terminated.
	<-svc.terminate
}

func (svc *Service) awaitSignals() {
	// Receive signal.
	sig := <-svc.signals
	sigName := unix.SignalName(sig.(syscall.Signal))

	// Clear characters from interrupt signal.
	if sig == syscall.SIGINT {
		fmt.Print("\r")
	}

	svc.Logger.Info().Msg("Signal received: " + sigName)
	svc.Logger.Info().Msg("Terminating ...")

	// Terminate process.
	svc.terminate <- true
}

// TODO: Create convenience function that handler errors
// by logging them. This allows to reduce the overhead
// on the business logic side.
