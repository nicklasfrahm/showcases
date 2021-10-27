package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/helmet/v2"

	"github.com/nicklasfrahm/showcases/pkg/broker"
	"github.com/nicklasfrahm/virtual-white-board/pkg/gateway"
	"github.com/nicklasfrahm/virtual-white-board/pkg/middleware"
	"github.com/nicklasfrahm/virtual-white-board/pkg/service"
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

	// Obtain configuration.
	port := os.Getenv("PORT")
	if port == "" {
		svc.Logger.Fatal().Msgf("Missing required environment variable: PORT")
	}

	svc.UseBroker(&broker.NATSOptions{}).UseGateway(&gateway.FiberConfig{})

	// Create router.
	app := fiber.New(fiber.Config{
		ErrorHandler:          middleware.Error(),
		DisableStartupMessage: true,
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
	app.Use(middleware.RedirectSlashes())
	app.Use(middleware.ContentType(fiber.MIMEApplicationJSONCharsetUTF8))

	// Mount gateway handler.
	app.Use(gateway.HTTP(svc))

	// Configure fallback route.
	app.Use(middleware.NotFound())

	svc.Logger.Info().Msg("Service online: " + port + "/tcp")
	if err := app.Listen(":" + port); err != nil {
		svc.Logger.Fatal().Err(err).Msg("Service failed")
	}

	// Wait until error occurs or signal is received.
	svc.Listen()
}
