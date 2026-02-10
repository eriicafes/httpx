// Package httperrors provides a structured way to handle HTTP errors with status codes and details.
package httperrors

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
)

// HTTPError represents an error that can be returned as an HTTP response.
type HTTPError interface {
	error

	// Message returns the user-facing error message.
	Message() string

	// StatusCode returns the HTTP status code for this error.
	StatusCode() int

	ErrorData
}

// ErrorData provides a method to retrieve additional metadata about an error.
type ErrorData interface {
	ErrorData() map[string]any
}

// Fields represents arbitrary custom fields that can be attached to an HTTPError.
// Unlike Details and Code which have fixed keys, Field allows you to add any top-level keys.
//
// Example:
//
//	fields := httperrors.Fields{
//		"request_id": "abc123",
//		"trace_id":   "xyz789",
//	}
//	err := httperrors.New("Internal error", http.StatusInternalServerError, fields)
type Fields map[string]any

func (f Fields) ErrorData() map[string]any {
	return f
}

// Details represents field-level validation error metadata that can be attached to an HTTPError.
// Useful for validation errors where specific fields have specific error messages.
//
// Example:
//
//	details := httperrors.Details{
//		"email":    "must be a valid email address",
//		"password": "must be at least 8 characters",
//	}
//	err := httperrors.New("Invalid input", http.StatusBadRequest, details)
type Details map[string]string

func (d Details) ErrorData() map[string]any {
	return map[string]any{"details": d}
}

// Code represents an error code that can be attached to an HTTPError.
// Error codes are usually a limited set defined as constants.
//
// Example:
//
//	const (
//		CodeNotFound     = httperrors.Code("NOT_FOUND")
//		CodeValidation   = httperrors.Code("VALIDATION_ERROR")
//		CodeUnauthorized = httperrors.Code("UNAUTHORIZED")
//	)
//
//	err := httperrors.New("User not found", http.StatusNotFound, CodeNotFound)
type Code string

func (c Code) ErrorData() map[string]any {
	return map[string]any{"code": c}
}

type httpError struct {
	err        error
	message    string
	statusCode int
	data       map[string]any
}

func (e *httpError) Error() string {
	if e.err == nil {
		return e.message
	}
	err := e.err.Error()
	if e.message == err {
		return e.message
	}
	return fmt.Sprintf("%s: %s", e.message, err)
}

func (e *httpError) Unwrap() error             { return e.err }
func (e *httpError) Message() string           { return e.message }
func (e *httpError) StatusCode() int           { return e.statusCode }
func (e *httpError) ErrorData() map[string]any { return e.data }

// New creates a new HTTPError with the given message, status code, and additional data.
//
// Example:
//
//	err := httperrors.New("Invalid input", http.StatusBadRequest, httperrors.Details{
//		"email": "must be a valid email address",
//		"password": "must be at least 8 characters",
//	})
func New(message string, statusCode int, data ...ErrorData) HTTPError {
	return &httpError{nil, message, statusCode, mergeErrorData(data)}
}

// Report converts a standard error into an HTTPError with status code, and additional data.
// The error's Error() message is used as the HTTP error message.
//
// Example:
//
//	if err := validator.Validate(input); err != nil {
//		return httperrors.Report(err, http.StatusBadRequest, httperrors.Details{
//			"username": "required field",
//		})
//	}
func Report(err error, statusCode int, data ...ErrorData) HTTPError {
	return &httpError{err, err.Error(), statusCode, mergeErrorData(data)}
}

// Wrap wraps an error into an HTTPError with a custom message, status code, and additional data.
// If err is already an HTTPError, its message, statusCode, and data will be used
// as fallbacks when the provided values are empty.
//
// Example:
//
//	if err := validateUser(input); err != nil {
//		return httperrors.Wrap(err, "Validation failed", http.StatusBadRequest, httperrors.Details{
//			"email": "already in use",
//		})
//	}
func Wrap(err error, message string, statusCode int, data ...ErrorData) HTTPError {
	mergedData := mergeErrorData(data)
	var herr HTTPError
	if errors.As(err, &herr) {
		if message == "" {
			message = herr.Message()
		}
		if statusCode == 0 {
			statusCode = herr.StatusCode()
		}
		if mergedData == nil {
			mergedData = herr.ErrorData()
		}
	}
	return &httpError{err, message, statusCode, mergedData}
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

func mergeErrorData(data []ErrorData) map[string]any {
	if len(data) == 0 {
		return nil
	}
	base := data[0].ErrorData()
	for _, d := range data[1:] {
		src := d.ErrorData()
		if src == nil {
			continue
		}
		if base == nil {
			base = src
			continue
		}
		maps.Copy(base, src)
	}
	return base
}

// Parse returns the message, status code, and data with defaults for empty values.
// The returned data map is always non-nil and includes an "error" field containing the message,
// without overriding any existing "error" field.
//
// This function is intended for use in application code where an error response needs to be sent.
// It accepts any error (HTTPError or standard error) and returns usable values for an HTTP response.
//
// Examples:
//
// Using with an HTTPError:
//
//	err := httperrors.New("Invalid email", http.StatusBadRequest, httperrors.Details{
//		"email": "must be a valid email address",
//	})
//	message, status, data := httperrors.Parse(err)
//	// message: "Invalid email"
//	// status: 400
//	// data: {"error": "Invalid email", "details": {"email": "must be a valid email address"}}
//
// Using with a standard error:
//
//	err := errors.New("database connection failed")
//	message, status, data := httperrors.Parse(err)
//	// message: "Something went wrong"
//	// status: 500
//	// data: {"error": "Something went wrong"}
//
// Sending a JSON error response:
//
//	if err != nil {
//		message, status, data := httperrors.Parse(err)
//		w.WriteHeader(status)
//		json.NewEncoder(w).Encode(data)
//		return
//	}
//
// Custom error field preservation:
//
//	err := httperrors.New("Operation failed", http.StatusBadRequest, httperrors.Fields{
//		"error": map[string]any{"code": "CUSTOM_ERROR", "message": "Custom error format"},
//	})
//	_, _, data := httperrors.Parse(err)
//	// data: {"error": {"code": "CUSTOM_ERROR", "message": "Custom error format"}}
//	// The existing "error" field is preserved, not overridden
func Parse(e error) (message string, statusCode int, data map[string]any) {
	if herr, ok := e.(HTTPError); ok {
		message = herr.Message()
		statusCode = herr.StatusCode()
		data = herr.ErrorData()
	}

	if message == "" {
		message = "Something went wrong"
	}
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	if data == nil {
		data = make(map[string]any)
	}
	if data["error"] == nil {
		data["error"] = message
	}
	return
}
