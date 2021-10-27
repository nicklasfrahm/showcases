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
	options *HTTPOptions
	app     *fiber.App
}

func NewHTTP(opts *HTTPOptions) service.Gateway {
	// TODO: Create a PreEndpoint and PostEndpoint method for better readability.

	// Create new fiber app.
	app := fiber.New(fiber.Config{
		ErrorHandler:          MiddlewareError(),
		DisableStartupMessage: true,
		Prefork:               opts.Prefork,
	})

	// Load middlewares.
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowHeaders:     "Accept,Authorization,Content-Type,X-CSRF-Token",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowCredentials: true,
		MaxAge:           600,
	}))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))
	app.Use(MiddlewareRedirectSlashes())
	app.Use(MiddlewareContentType(fiber.MIMEApplicationJSONCharsetUTF8))

	return &HTTP{
		options: opts,
		app:     app,
	}
}

func (g *HTTP) Bind(svc *service.Service) {
	g.service = svc
}

func (g *HTTP) Route(endpoint string, requestHandler service.RequestHandler) {
	g.app.Use(endpoint, func(c *fiber.Ctx) error {
		return requestHandler(&service.Request{
			Ctx:     c,
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
