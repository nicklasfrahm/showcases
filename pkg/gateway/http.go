package gateway

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/helmet/v2"

	"github.com/nicklasfrahm/showcases/pkg/service"
)

type HTTPOptions struct {
	Prefork bool
	Port    string
}

type HTTP struct {
	service *service.Service

	options *Options
	app     *fiber.App
}

// NewHTTP creates and configures a new HTTP gateway.
// For more information about the protocol translation
// refer to: https://docs.mykil.io/ecosystem/concepts.html
func NewHTTP(options ...Option) service.Gateway {
	opts := GetDefaultOptions()
	for _, opt := range options {
		if opt != nil {
			if err := opt(&opts); err != nil {
				return nil
			}
		}
	}
	return &HTTP{
		options: &opts,
	}
}

func (g *HTTP) Bind(svc *service.Service) {
	g.service = svc

	// Create new fiber app.
	g.app = fiber.New(fiber.Config{
		ErrorHandler:          MiddlewareError(),
		DisableStartupMessage: true,
		Prefork:               g.options.Prefork,
	})
	g.app.Use(recover.New())
	g.app.Use(helmet.New())
	g.app.Use(cors.New(cors.Config{
		AllowHeaders:     "Accept,Authorization,Content-Type,X-CSRF-Token",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowCredentials: true,
		MaxAge:           600,
	}))
	g.app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))
	g.app.Use(MiddlewareRedirectSlashes())
	g.app.Use(MiddlewareContentType(fiber.MIMEApplicationJSONCharsetUTF8))
}

func (g *HTTP) Route(requestHandler service.RequestHandler) {
	g.app.Use(func(c *fiber.Ctx) error {
		return requestHandler(&service.Request{
			Context: c,
			Service: g.service,
		})
	})
}

func (g *HTTP) Listen() {
	// Configure fallback route.
	g.app.Use(MiddlewareNotFound())

	// Check if the port is valid.
	if g.options.Port == "" {
		g.service.Logger.Warn().Msg("Missing environment variable: PORT")
		g.options.Port = "8080"
		g.service.Logger.Warn().Msgf("Using default port: %s/tcp", g.options.Port)
	}

	g.service.Logger.Info().Msgf("Gateway online: %s/tcp", g.options.Port)
	if err := g.app.Listen(":" + g.options.Port); err != nil {
		g.service.Logger.Fatal().Err(err).Msg("Failed running gateway")
	}
}

func (g *HTTP) Context(r *service.Request) *fiber.Ctx {
	// Do not check for successful type assertion. This should
	// fail dramatically if the type is wrong.
	return r.Context.(*fiber.Ctx)
}
