package openapi

import (
	"net/http"

	"github.com/eriicafes/httpx"
	"github.com/eriicafes/httpx/openapi/doc"
)

// WithRouter wraps mux with an embedded OpenAPI Router configured by opts.
// Use UseRouter to retrieve the Router and register OpenAPI operations and route handlers.
func WithRouter(mux httpx.ServeMux, opts ...doc.Option) httpx.Mux {
	return &openapiMux{mux, NewRouter(opts...)}
}

type openapiMux struct {
	mux    httpx.ServeMux
	router *Router
}

func (m *openapiMux) Mux() httpx.ServeMux {
	return m.mux
}

func (m *openapiMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

func (m *openapiMux) Handle(pattern string, handler http.Handler) {
	m.mux.Handle(pattern, handler)
}

func (m *openapiMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.Handle(pattern, http.HandlerFunc(handler))
}

func (m *openapiMux) Route(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	m.Handle(pattern, httpx.Handler(m, handler))
}
