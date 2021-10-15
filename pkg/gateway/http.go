package gateway

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/nicklasfrahm/virtual-white-board/pkg/errs"
	"github.com/nicklasfrahm/virtual-white-board/pkg/service"
)

var (
	requestOne = map[string]string{
		http.MethodGet:    "read",
		http.MethodPut:    "update",
		http.MethodDelete: "delete",
	}
	requestMany = map[string]string{
		http.MethodGet:  "find",
		http.MethodPost: "create",
	}
)

type Meta struct {
	Subject string
	Verb    string
}

func (m *Meta) Type() string {
	return m.Subject + "." + m.Verb
}

// HTTP returns a middlware that converts.
func HTTP(svc *service.Service) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// Convert HTTP path to NATS subject.
		meta, err := metaFromPath(c)
		if err != nil {
			return err
		}

		// Parse body.
		var body interface{}
		if c.Method() == http.MethodPost || c.Method() == http.MethodPut {
			if err := c.BodyParser(&body); err != nil {
				svc.Logger.Warn().Msg(err.Error())
				return err
			}
		}

		event := cloudevents.NewEvent()
		event.SetID(uuid.NewString())
		event.SetSource("gateway-http")
		event.SetType(meta.Type())
		event.SetData(fiber.MIMEApplicationJSONCharsetUTF8, body)

		// Encode cloud event.
		encodedEvent, err := json.Marshal(event)
		if err != nil {
			return err
		}

		msg, err := svc.Broker.Request(meta.Subject, encodedEvent, 10*time.Millisecond)
		if err != nil {
			return errs.InvalidService
		}

		// TODO: Set HTTP status based on service response.

		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		return c.Send(msg.Data)
	}
}

// TODO: Test for edge cases, such as dots in the path.
func metaFromPath(c *fiber.Ctx) (*Meta, error) {
	// Convert resource path to NATS subject.
	resourceSubject := strings.ReplaceAll(c.Path()[1:], "/", ".")

	// Check if the path describes a specific resource or a resource list, assuming the scheme is /resource/:rid/subresource/:srid.
	method := c.Method()
	if strings.Count(resourceSubject, ".")%2 == 0 {
		// Resource lists do not support PUT or DELETE methods.
		if method == http.MethodPut || method == http.MethodDelete {
			return nil, errs.InvalidEndpoint
		}
		return &Meta{Subject: resourceSubject, Verb: requestMany[method]}, nil
	}

	// Specific resources do not support the POST method.
	if method == http.MethodPost {
		return nil, errs.InvalidEndpoint
	}
	return &Meta{Subject: resourceSubject, Verb: requestOne[method]}, nil
}
