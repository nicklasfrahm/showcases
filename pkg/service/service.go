package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
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
	Logger *zerolog.Logger
	Broker *nats.Conn

	config    Config
	signals   chan os.Signal
	terminate chan bool
}

// New returns a new service for the given configuration.
func New(config Config) *Service {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	return &Service{
		// Configure logger.
		Logger: &log.Logger,

		config:    config,
		signals:   make(chan os.Signal, 1),
		terminate: make(chan bool, 1),
	}
}

func (svc *Service) UseBroker(uri string) {
	if uri == "" {
		svc.Logger.Fatal().Msgf("Broker URI missing")
	}

	options := []nats.Option{
		nats.Name(svc.config.Name),
		nats.Timeout(1 * time.Second),
		nats.PingInterval(5 * time.Second),
		nats.MaxPingsOutstanding(6),
	}

	// TODO: Remove this once credentials are used.
	svc.Logger.Info().Msgf("Connecting to broker: %s", uri)
	conn, err := nats.Connect(uri, options...)
	if err != nil {
		svc.Logger.Fatal().Msg(err.Error())
	}

	svc.Broker = conn
}

func (svc *Service) Listen() {
	// Subscribe to OS signals and asynchronously await them in goroutine.
	signal.Notify(svc.signals, syscall.SIGINT, syscall.SIGTERM)
	go svc.awaitSignal()

	// Log basic service information.
	svc.Logger.Info().Msg("Service: " + svc.config.Name)
	svc.Logger.Info().Msg("Version: " + svc.config.Version)

	// Block until terminated.
	<-svc.terminate
}

func (svc *Service) awaitSignal() {
	// Receive signal.
	sig := <-svc.signals
	sigName := unix.SignalName(sig.(syscall.Signal))

	// Clear characters from interrupt signal.
	if sig == syscall.SIGINT {
		fmt.Print("\r")
	}
	svc.Logger.Info().Msg("Terminating due to signal: " + sigName)

	// Terminate process.
	svc.terminate <- true
}
