package httpx

import (
	"net/http"
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

// HandlerFunc is a handler func that returns an error
type HandlerFunc func(http.ResponseWriter, *http.Request) error

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

func (f ErrorHandlerFunc) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	f(w, r, err)
}

type ErrorHandler interface {
	HandleError(http.ResponseWriter, *http.Request, error)
}

// MuxErrorHandler returns the error handler for mux.
// If mux does not implement the HandleError interface the returned
// error handler will write a default error response.
func MuxErrorHandler(mux ServeMux) ErrorHandler {
	for {
		switch m := mux.(type) {
		case ErrorHandler:
			return m
		case interface{ SubMux() ServeMux }:
			mux = m.SubMux()
		default:
			return ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			})
		}
	}
}

// ApplyMuxErrorHandler wraps an error-returning handler with automatic error handling.
// The error handler is retrieved from the mux using MuxErrorHandler.
func ApplyMuxErrorHandler(mux ServeMux, handler func(http.ResponseWriter, *http.Request) error) http.Handler {
	errorHandler := MuxErrorHandler(mux)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			errorHandler.HandleError(w, r, err)
		}
	})
}
