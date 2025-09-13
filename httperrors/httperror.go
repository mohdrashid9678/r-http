package httperrors

import "fmt"

// HTTPError is a standard error type.
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error %d: %s", e.StatusCode, e.Message)
}

func NewBadRequest(message string) *HTTPError {
	return &HTTPError{StatusCode: 400, Message: message}
}

func NewNotFound(resource string) *HTTPError {
	return &HTTPError{StatusCode: 404, Message: fmt.Sprintf("Resource '%s' not found", resource)}
}

func NewInternalServerError(message string) *HTTPError {
	return &HTTPError{StatusCode: 500, Message: message}
}
