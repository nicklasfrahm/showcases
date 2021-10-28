package main

import (
	"os"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/nicklasfrahm/showcases/pkg/broker"
	"github.com/nicklasfrahm/showcases/pkg/mail"
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

	// Fetch credentials from environment.
	mailFrom := os.Getenv("MAIL_FROM")
	sparkpostAPIKey := os.Getenv("SPARKPOST_API_KEY")
	sparkpostHTTPURI := os.Getenv("SPARKPOST_HTTP_URI")
	sendgridAPIKey := os.Getenv("SENDGRID_API_KEY")
	sendgridHTTPURI := os.Getenv("SENDGRID_HTTP_URI")

	// Configure email providers.
	mailers := make(map[string]mail.Mailer)

	// Set up mailers.
	mailers["sendgrid-http"] = mail.NewSendgridHTTP(&mail.Config{
		URI:     sendgridHTTPURI,
		APIKey:  sendgridAPIKey,
		Logger:  svc.Logger,
		Timeout: 1 * time.Second,
	})
	mailers["sparkpost-http"] = mail.NewSparkpostHTTP(&mail.Config{
		URI:     sparkpostHTTPURI,
		APIKey:  sparkpostAPIKey,
		Logger:  svc.Logger,
		From:    mailFrom,
		Timeout: 1 * time.Second,
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

	svc.BrokerEndpoint("v1.services.mail.providers.find", func(ctx *service.Context) {
		// Fetch information about mail providers.
		mailProviders := make([]mail.MailProvider, len(mailers))
		i := 0
		for _, mailer := range mailers {
			// This should be more performant than using append().
			mailProviders[i] = mailer.MailProvider()
			i += 1
		}

		// Send reply. Please note that the source is an opaque string
		// that is used by the broker implementation to perform routing.
		if err := ctx.Service.Broker.Publish(ctx.Cloudevent.Source(), mailProviders); err != nil {
			ctx.Service.Logger.Error().Err(err).Msgf("Failed to respond")
		}
		// Broadcast event.
		if err := ctx.Service.Broker.Publish("v1.services.mail.providers.found", mailProviders); err != nil {
			ctx.Service.Logger.Error().Err(err).Msgf("Failed to broadcast")
		}
	})

	svc.BrokerEndpoint("v1.mails.create", func(ctx *service.Context) {
		// Parse cloudevent and marshal it into a struct.
		mail := new(mail.Mail)
		if err := ctx.Cloudevent.DataAs(mail); err != nil {
			// TODO: Improve error handling by sending appropriate error code
			// to gateway like as 400 or 422, because the validation failed.
			// This will just silently fail and cause the service to return
			// error code 503, which is not very descriptive.
			return
		}

		sent := false
		for _, mailer := range mailers {
			// Check if provider is disabled.
			if !mailer.MailProvider().Disabled {
				// Attempt to send email.
				err := mailer.Send(mail)
				if err == nil {
					// Sucessfully sent email. Don't retry.
					sent = true
					break
				}

				// Display warning message upon failed delivery attempt.
				ctx.Service.Logger.Warn().Err(err).Msgf("Failed to send mail")
			}
		}

		if !sent {
			// TODO: See earlier to do comment. Bad Nicklas, go fix!
			return
		}

		// Send reply.
		if err := ctx.Service.Broker.Publish(ctx.Cloudevent.Source(), mail); err != nil {
			ctx.Service.Logger.Error().Err(err).Msgf("Failed to respond")
		}
		// Broadcast unsent email.
		if err := ctx.Service.Broker.Publish("v1.mails.unsent", mail); err != nil {
			ctx.Service.Logger.Error().Err(err).Msgf("Failed to broadcast")
		}
	})

	// Wait until error occurs or signal is received.
	svc.Start()
}
