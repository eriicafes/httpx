// Package httpx provides extended HTTP utilities for Go, extending net/http
// with error-returning handlers, middleware support, graceful shutdown, and
// composable mux wrappers.
//
// # Error-Returning Handlers
//
// The core feature is handlers that return errors instead of manually writing
// error responses:
//
//	mux := httpx.New()
//	mux.Route("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) error {
//	    user, err := getUser(r.PathValue("id"))
//	    if err != nil {
//	        return httpx.InternalError(err, "Failed to retrieve user")
//	    }
//	    return httpx.Send(w, user)
//	})
//
// Errors are automatically converted to JSON responses with appropriate status codes:
//   - Regular errors: 400 Bad Request
//   - Internal errors (httpx.InternalError): 500 Internal Server Error (logged server-side)
//   - HTTP errors (httperrors.HTTPError): Custom status codes and details
//
// # Graceful Shutdown
//
// The ListenAndServe function provides automatic graceful shutdown on SIGINT/SIGTERM:
//
//	httpx.ListenAndServe(":8080", mux, &httpx.ServerConfig{
//	    ShutdownTimeout: 30 * time.Second,
//	})
//
// # Middleware
//
// Apply middleware to routes using the Use wrapper:
//
//	mux := httpx.Use(
//	    http.NewServeMux(),
//	    loggingMiddleware,
//	    authMiddleware,
//	)
//
// # Composable Mux Wrappers
//
// Chain multiple mux wrappers for layered functionality:
//
//	mux := httpx.New()
//	mux = httpx.NormalizeTrailingSlash(mux)
//	mux = httpx.Prefix(mux, "/api/v1")
//	mux = httpx.Use(mux, corsMiddleware)
//	mux = httpx.Fallback(mux, customErrorHandler)
//
// Available mux wrappers:
//   - Prefix: Add common prefix to routes
//   - Use: Apply middleware
//   - Fallback: Custom error handling
//   - NormalizeTrailingSlash: Handle routes with/without trailing slashes
//
// # Request and Response Helpers
//
// Simplified JSON encoding and decoding:
//
//	// Read request body
//	user, err := httpx.Read[User](r)
//
//	// Send response
//	httpx.Send(w, user)
//	httpx.SendStatus(w, http.StatusCreated, user)
//
// # Subpackages
//
// The httpx module includes specialized subpackages:
//
//   - httperrors: Structured HTTP error handling with status codes
//   - contextkey: Type-safe context key management
//   - session: Session-based authentication with cookies and flash messages
package httpx

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/eriicafes/httpx/httperrors"
)

// ServeMux is the generic interface for route registering mux.
type ServeMux interface {
	http.Handler
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

// Mux is a specialized mux that can register routes with handlers that return an error.
type Mux interface {
	ServeMux
	Route(pattern string, handler func(http.ResponseWriter, *http.Request) error)
}

func New() Mux {
	return Use(http.NewServeMux())
}

// HandlerFunc allows the use of handler functions that return an error
// as HTTP handlers. HandlerFunc(f) is a [http.Handler] that calls f.
// If returned error is non-nil, it handles the error using the default error handler.
//
// The default error handler replies to the request with a JSON response
// of the error message and an [http.StatusInternalServerError] code.
// It also supports [httperrors.HTTPError] and [InternalError].
//
// For better control over error responses:
//   - Use [InternalError] to wrap errors with user-friendly messages
//   - Use the httperrors package for structured errors with custom status codes
type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f(w, r); err != nil {
		defaultErrorHandler(w, r, err)
	}
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

func (f ErrorHandlerFunc) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	f(w, r, err)
}

type ErrorHandler interface {
	HandleError(http.ResponseWriter, *http.Request, error)
}

// MuxErrorHandler returns the error handler for mux.
// If mux does not implement the [ErrorHandler] interface,
// the default error handler is returned.
//
// The default error handler replies to the request with a JSON response
// of the error message and an [http.StatusInternalServerError] code.
// It also supports [httperrors.HTTPError] and [InternalError].
func MuxErrorHandler(mux ServeMux) ErrorHandler {
	for {
		switch m := mux.(type) {
		case ErrorHandler:
			return m
		case interface{ SubMux() ServeMux }:
			mux = m.SubMux()
		default:
			return ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request, err error) {
				defaultErrorHandler(w, r, err)
			})
		}
	}
}

// ApplyMuxErrorHandler wraps an error-returning handler with the mux's error handling.
// The error handler is retrieved from the mux using [MuxErrorHandler].
func ApplyMuxErrorHandler(mux ServeMux, handler func(http.ResponseWriter, *http.Request) error) http.Handler {
	errorHandler := MuxErrorHandler(mux)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			errorHandler.HandleError(w, r, err)
		}
	})
}

func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	h := w.Header()
	h.Del("Content-Length")
	h.Set("Content-Type", "application/json")
	h.Set("X-Content-Type-Options", "nosniff")

	// Handle HTTPError
	if httpErr, ok := httperrors.Unwrap(err); ok {
		_, statusCode, data := httpErr.HTTPError()

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(data)
		return
	}

	// Default error handling
	var internalErr *internalError
	if errors.As(err, &internalErr) {
		// Log underlying error for internal errors
		log.Printf("[ERROR] %s %s: %v", r.Method, r.URL.Path, internalErr.err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"error": err.Error(),
	})
}

// InternalError wraps an error with a user-friendly message for client responses.
// If message is empty, it defaults to "Something went wrong".
//
// When used with the default error handler, it logs the underlying error
// (with request method and path) while returning only the user-friendly message to the client.
// This prevents internal implementation details from leaking while maintaining debug visibility.
//
//	user, err := getUser(id)
//	if err != nil {
//	    // Logs: "[ERROR] GET /api/users/123: sql: no rows in result set"
//	    // Returns to client: {"error": "Failed to retrieve user"}
//	    return httpx.InternalError(err, "Failed to retrieve user")
//	}
func InternalError(err error, message string) error {
	if message == "" {
		message = "Something went wrong"
	}
	return &internalError{
		err:     err,
		message: message,
	}
}

// internalError wraps an error with a user-friendly message.
// The default error handler logs the underlying error with request details
// and returns the user-friendly message to clients.
type internalError struct {
	err     error
	message string
}

func (e *internalError) Error() string {
	return e.message
}

func (e *internalError) Unwrap() error {
	return e.err
}
