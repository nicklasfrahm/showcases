package gateway

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/nicklasfrahm/showcases/pkg/errs"
)

// ErrorResponse is the payload sent in the case of an error.
type ErrorResponse struct {
	Error errs.ServiceError `json:"error"`
}

// MiddlewareContentType is a middleware to set the "Content-Type" header.
func MiddlewareContentType(contentType string) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Request().Header.Set(fiber.HeaderContentType, contentType)
		return c.Next()
	}
}

// MiddlewareError returns a middleware to handler errors during requests.
func MiddlewareError() func(*fiber.Ctx, error) error {
	return func(c *fiber.Ctx, err error) error {
		// Handle known service error types.
		if svcErr, ok := err.(*errs.ServiceError); ok {
			return c.Status(svcErr.Status).JSON(ErrorResponse{
				Error: *svcErr,
			})
		}

		// Return default error.
		defErr := errs.UnexpectedError
		return c.Status(defErr.Status).JSON(ErrorResponse{
			Error: *defErr,
		})
	}
}

// MiddlewareNotFound returns a middlware for endpoints that do not exist.
func MiddlewareNotFound() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return errs.InvalidEndpoint
	}
}

// MiddlewareRedirectSlashes slashes redirects routes with a trailing slash
// to the same route without the trailing slash.
func MiddlewareRedirectSlashes() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		if len(path) > 1 && path[len(path)-1] == '/' {
			segments := strings.Split(c.OriginalURL(), "?")
			query := ""
			if len(segments) == 2 {
				query = "?" + segments[1]
			}
			return c.Redirect(path[:len(path)-1]+query, http.StatusMovedPermanently)
		}

		return c.Next()
	}
}
