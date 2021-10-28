package errs

import (
	"net/http"
)

var (
	InvalidJSON        = NewServiceError(400, "Invalid JSON")
	MissingCredentials = NewServiceError(401, "Missing Credentials")
	InvalidCredentials = NewServiceError(403, "Invalid Credentials")
	InvalidEndpoint    = NewServiceError(404, "Invalid Endpoint")
	UnexpectedError    = NewServiceError(500, "Unexpected Error")
	InvalidService     = NewServiceError(503, "Invalid Service")
)

// ServiceError describes required information in case of an error.
type ServiceError struct {
	Title   string `json:"title"`
	Status  int    `json:"status"`
	Message string `json:"message"`

	Err error `json:"-"`
}

func (se *ServiceError) Error() string {
	return se.Message
}

func NewServiceError(status int, message string) *ServiceError {
	return &ServiceError{
		Title:   http.StatusText(status),
		Status:  status,
		Message: message,
	}
}
