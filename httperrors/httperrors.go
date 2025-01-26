package httperrors

import (
	"errors"
	"net/http"
)

type HTTPError interface {
	error
	Unwrap() error
	HTTPError() (statusCode int, message string, details Details)
}

type Details = map[string]string

// New returns a new HTTPError with status code.
// The error message is set to the msg string.
func New(msg string, statusCode int) HTTPError {
	return &httpError{errors.New(msg), statusCode, nil}
}

// Wrap wraps an error into an HTTPError with status code.
// If err is an HTTPError it's details will be preserved.
// The error message is set to the wrapped error's error message.
func Wrap(err error, statusCode int) HTTPError {
	var details Details
	if herr, ok := err.(HTTPError); ok {
		_, _, details = herr.HTTPError()
	}
	return WrapDetails(err, statusCode, details)
}

// WrapDetails wraps an error into an HTTPError with statusCode and details.
// The error message is set to the wrapped error's error message.
func WrapDetails(err error, statusCode int, details Details) HTTPError {
	return &httpError{err, statusCode, details}
}

type httpError struct {
	error
	statusCode int
	details    Details
}

func (e *httpError) Unwrap() error {
	return e.error
}

func (e *httpError) HTTPError() (int, string, Details) {
	statusCode := e.statusCode
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	return statusCode, e.error.Error(), e.details
}

// Opaque wraps an error into an HTTPError with status code while hiding it's error message.
// The error message is set to the msg string.
func Opaque(err error, msg string, statusCode int) error {
	return &opaqueError{&httpError{errors.New(msg), statusCode, nil}, err}
}

type opaqueError struct {
	HTTPError
	err error
}

func (e *opaqueError) Unwrap() error {
	return e.err
}
