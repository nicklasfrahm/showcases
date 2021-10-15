package err

import (
	"net/http"
)

var (
	InvalidEndpoint = NewServiceError(404, "Invalid Endpoint")
	UnexpectedError = NewServiceError(500, "Unexpected Error")
)

// ServiceError describes required information in case of an error.
type ServiceError struct {
	Title   string `json:"title"`
	Status  int    `json:"status"`
	Message string `json:"message"`

	Err error
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
