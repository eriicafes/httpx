// Package httperrors provides a structured way to handle HTTP errors with status codes and details.
package httperrors

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
)

// HTTPError represents an error that can be returned as an HTTP response.
type HTTPError interface {
	error
	// Message returns the user-facing error message.
	Message() string
	// StatusCode returns the HTTP status code for this error.
	StatusCode() int
	// Details returns additional metadata about the error.
	Details() Details
	// HTTPError returns the message, status code, and details with defaults for empty values.
	HTTPError() (message string, statusCode int, details Details)
}

// Details represents additional metadata that can be attached to an error.
type Details = map[string]string

type httpError struct {
	err        error
	message    string
	statusCode int
	details    Details
}

func (e *httpError) Error() string {
	if e.err == nil {
		return e.message
	}
	mssg := e.err.Error()
	if e.message == mssg {
		return e.message
	}
	return fmt.Sprintf("%s: %s", e.message, mssg)
}

func (e *httpError) Unwrap() error { return e.err }

func (e *httpError) Message() string  { return e.message }
func (e *httpError) StatusCode() int  { return e.statusCode }
func (e *httpError) Details() Details { return e.details }

func (e *httpError) HTTPError() (string, int, Details) {
	message := cmp.Or(e.message, "Something went wrong")
	statusCode := cmp.Or(e.statusCode, http.StatusInternalServerError)
	return message, statusCode, e.details
}

// New creates a new HTTPError with the given message and status code.
//
// Example:
//
//	err := httperrors.New("User not found", http.StatusNotFound)
func New(message string, statusCode int) HTTPError {
	return &httpError{nil, message, statusCode, nil}
}

// NewDetails creates a new HTTPError with the given message, status code, and additional details.
//
// Example:
//
//	err := httperrors.NewDetails("Invalid input", http.StatusBadRequest, httperrors.Details{
//		"email": "must be a valid email address",
//		"password": "must be at least 8 characters",
//	})
func NewDetails(message string, statusCode int, details Details) HTTPError {
	return &httpError{nil, message, statusCode, details}
}

// Report converts a standard error into an HTTPError.
// The error's Error() message is used as the HTTP error message.
//
// Example:
//
//	if err := db.QueryRow(...); err != nil {
//		return httperrors.Report(err, http.StatusInternalServerError)
//	}
func Report(err error, statusCode int) HTTPError {
	return &httpError{err, err.Error(), statusCode, nil}
}

// ReportDetails converts a standard error into an HTTPError with additional details.
// The error's Error() message is used as the HTTP error message.
//
// Example:
//
//	if err := validator.Validate(input); err != nil {
//		return httperrors.ReportDetails(err, http.StatusBadRequest, httperrors.Details{
//			"username": "required field",
//		})
//	}
func ReportDetails(err error, statusCode int, details Details) HTTPError {
	return &httpError{err, err.Error(), statusCode, details}
}

// Wrap wraps an error into an HTTPError with a custom message and status code.
// If err is already an HTTPError, its message, statusCode, and details will be used
// as fallbacks when the provided values are empty.
//
// Example:
//
//	if err := processPayment(); err != nil {
//		return httperrors.Wrap(err, "Payment processing failed", http.StatusBadGateway)
//	}
func Wrap(err error, message string, statusCode int) HTTPError {
	return WrapDetails(err, message, statusCode, nil)
}

// WrapDetails wraps an error into an HTTPError with a custom message, status code, and details.
// If err is already an HTTPError, its message, statusCode, and details will be used
// as fallbacks when the provided values are empty.
//
// Example:
//
//	if err := validateUser(input); err != nil {
//		return httperrors.WrapDetails(err, "Validation failed", http.StatusBadRequest, httperrors.Details{
//			"email": "already in use",
//		})
//	}
func WrapDetails(err error, message string, statusCode int, details Details) HTTPError {
	var herr HTTPError
	if errors.As(err, &herr) {
		if statusCode == 0 {
			statusCode = herr.StatusCode()
		}
		if message == "" {
			message = herr.Message()
		}
		if details == nil {
			details = herr.Details()
		}
	}
	return &httpError{err, message, statusCode, details}
}

// Unwrap checks if an error is or wraps an HTTPError.
// It returns the HTTPError and true if found, otherwise nil and false.
//
// Example:
//
//	if httpErr, ok := httperrors.Unwrap(err); ok {
//		log.Printf("HTTP error: %d - %s", httpErr.StatusCode(), httpErr.Message())
//	}
func Unwrap(err error) (HTTPError, bool) {
	var herr HTTPError
	ok := errors.As(err, &herr)
	return herr, ok
}
