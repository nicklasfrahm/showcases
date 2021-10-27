package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
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
	mailers["sparkpost-http"] = mail.NewSparkpostHTTP(&mail.Config{
		URI:     sparkpostHTTPURI,
		APIKey:  sparkpostAPIKey,
		Logger:  svc.Logger,
		From:    mailFrom,
		Timeout: 1 * time.Second,
	})
	mailers["sendgrid-http"] = mail.NewSendgridHTTP(&mail.Config{
		URI:     sendgridHTTPURI,
		APIKey:  sendgridAPIKey,
		Logger:  svc.Logger,
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

	svc.BrokerEndpoint("v1.mail.create", func(m *service.Message) {
		// TODO: Parse cloudevent and marshal it into a struct.
		req := cloudevents.NewEvent()
		fmt.Println(string(*m.Data))
		err := json.Unmarshal(*m.Data, &req)
		if err != nil {
			m.Service.Logger.Warn().Err(err).Msgf("Failed to send mail")
			return
		}

		mail := &mail.Mail{}

		res := cloudevents.NewEvent()
		res.SetID(uuid.NewString())
		res.SetSource("mail")
		for _, mailer := range mailers {
			// Check if provider is disabled.
			if !mailer.MailProvider().Disabled {
				// Attempt to send email.
				if err = mailer.Send(mail); err == nil {
					// Sucessfully sent email. Don't retry.
					res.SetType("v1.mail.created")
					res.SetDataContentType(cloudevents.ApplicationJSON)
					res.SetData(mail)

					// Do not attempt to use other providers.
					break
				}

				// Display warning message upon failed delivery attempt.
				m.Service.Logger.Warn().Err(err).Msgf("Failed to send mail")
				// Reset error.
				err = nil
			}
		}

		// No attempt was sucessful.
		endpointUnsent := "v1.mail.unsent"
		if err != nil {
			res.SetType(endpointUnsent)
			res.SetDataContentType(cloudevents.ApplicationJSON)
			res.SetData(mail)
		}

		// Encode cloud event.
		encodedEvent, err := json.Marshal(res)
		if err != nil {
			m.Service.Logger.Warn().Err(err).Msgf("Failed to encode response")
			return
		}

		// Send reply.
		if err := m.Service.Broker.Publish(*m.Reply, encodedEvent); err != nil {
			m.Service.Logger.Error().Err(err).Msgf("Failed to send response")
		}
		// Broadcast unsent email.
		if err := m.Service.Broker.Publish(endpointUnsent, encodedEvent); err != nil {
			m.Service.Logger.Error().Err(err).Msgf("Failed to reply: %s")
		}
	})

	// Wait until error occurs or signal is received.
	svc.Start()
}
