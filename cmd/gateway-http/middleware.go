package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nicklasfrahm/showcases/pkg/errs"
	"github.com/nicklasfrahm/showcases/pkg/service"
)

const (
	LocalsChannel = "channel"
	LocalsType    = "type"
)

func NormalizeProtoToChannel() service.RequestHandler {
	return func(r *service.Request) error {
		ctx := r.Context.(*fiber.Ctx)
		method := ctx.Method()
		path := ctx.Path()

		// Convert HTTP path to canonical channel.
		var channel, eventType string
		segments := strings.Split(path[1:], "/")
		resource := strings.Join(segments, ".")
		// Check if the path describes a specific resource or a resource list.
		// This assumes that thescheme is: /resource/:rid/subresource/:srid.
		if len(segments)%2 != 0 {
			// Resource lists only support GET or POST methods.
			if method != http.MethodGet && method != http.MethodPost {
				return errs.InvalidEndpoint
			}
			channel = fmt.Sprintf("%s.%s", resource, mapListMethod[method])
			eventType = fmt.Sprintf("%s.%s", segments[len(segments)-1], mapListMethod[method])
		} else {
			// Specific resources only support GET, PUT and DELETE methods.
			if method != http.MethodGet && method != http.MethodPut && method != http.MethodDelete {
				return errs.InvalidEndpoint
			}
			channel = fmt.Sprintf("%s.%s", resource, mapMethod[method])
			if len(segments) == 0 {
				eventType = fmt.Sprintf("root.%s", mapListMethod[method])
			} else {
				eventType = fmt.Sprintf("%s.%s", segments[len(segments)-2], mapListMethod[method])
			}
		}

		// Persist information to request.
		ctx.Locals(LocalsChannel, channel)
		ctx.Locals(LocalsType, eventType)

		return ctx.Next()
	}
}

func AuthN(users map[string]string) service.RequestHandler {
	return func(r *service.Request) error {
		ctx := r.Context.(*fiber.Ctx)

		// Decode information from request header.
		authHeader := ctx.Request().Header.Peek("Authorization")
		authHeaderSegments := strings.Split(string(authHeader), " ")

		// Check if credentials are present with the right authentication scheme.
		if len(authHeaderSegments) != 2 {
			return errs.MissingCredentials
		}
		authScheme := authHeaderSegments[0]
		if strings.ToLower(authScheme) != "basic" {
			return errs.InvalidCredentials
		}

		// Decode the credentials.
		authCreds, err := base64.StdEncoding.DecodeString(authHeaderSegments[1])
		if err != nil {
			return errs.InvalidCredentials
		}

		credentials := strings.SplitN(string(authCreds), ":", 2)
		if len(credentials) != 2 {
			return errs.InvalidCredentials
		}

		user := credentials[0]
		pass := credentials[1]
		if users[user] != pass {
			return errs.InvalidCredentials
		}
		return ctx.Next()
	}
}

func DispatchToChannel() service.RequestHandler {
	return func(r *service.Request) error {
		ctx := r.Context.(*fiber.Ctx)

		// Fetch body from request.
		var body interface{}
		if ctx.Method() == http.MethodPost || ctx.Method() == http.MethodPut {
			// Parse body.
			if err := ctx.BodyParser(&body); err != nil {
				return errs.InvalidJSON
			}
		}

		channel := ctx.Locals(LocalsChannel).(string)
		res, err := r.Service.Broker.Request(channel, body)
		if err != nil {
			return errs.InvalidService
		}

		// TODO: Check the type of the event.
		_ = res.Cloudevent.Type()

		// TODO: Set HTTP status based on service response.
		ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		return ctx.Send(res.Cloudevent.Data())
	}
}
