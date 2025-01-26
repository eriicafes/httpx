package httpx

import (
	"net/http"
)

// ServeMux is the generic interface for route registering mux.
type ServeMux interface {
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

// Mux is a specialized mux that can register routes with handlers that return an error.
type Mux interface {
	ServeMux
	Route(pattern string, handler func(http.ResponseWriter, *http.Request) error)
}
