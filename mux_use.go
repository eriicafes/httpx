package httpx

import (
	"net/http"
	"slices"
)

// Middleware is a function that accepts a handler and returns another handler.
// Middlewares can decide to call or not to call the handler and they can run before or after the handler.
type Middleware func(handler http.Handler) http.Handler

// Use returns a Mux which applies middlewares to the handler.
// Use with an empty middlewares list can be used to convert a *http.ServeMux to a Mux.
func Use(mux ServeMux, middlewares ...Middleware) Mux {
	return &useMux{mux, middlewares}
}

type useMux struct {
	mux         ServeMux
	middlewares []Middleware
}

func (m *useMux) SubMux() ServeMux {
	return m.mux
}

func (m *useMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

func (m *useMux) Handle(pattern string, handler http.Handler) {
	for _, mh := range slices.Backward(m.middlewares) {
		handler = mh(handler)
	}
	m.mux.Handle(pattern, handler)
}

func (m *useMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.Handle(pattern, http.HandlerFunc(handler))
}

func (m *useMux) Route(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	m.Handle(pattern, ApplyMuxErrorHandler(m, handler))
}
