package httpx

import (
	"net/http"
	"strings"
)

// Prefix returns a Mux that registers routes under the given prefix.
//
// Prefix concatenates the prefix with each registered pattern and registers
// the resulting pattern on the underlying mux. Any HTTP method specified in
// the pattern (for example "GET /users") is preserved in the resulting pattern.
// No other parsing, normalization, or rewriting is performed.
//
// The prefix must start with '/'.
// The pattern may be empty ("") or start with '/'.
//
// Examples:
//
//	Prefix(mux, "/api").Handle("", h)
//	  → registers "/api"
//
//	Prefix(mux, "/api").Handle("/", h)
//	  → registers "/api/"
//
//	Prefix(mux, "/api").Handle("POST ", h)
//	  → registers "POST /api"
//
//	Prefix(mux, "/api").Handle("/users", h)
//	  → registers "/api/users"
//
//	Prefix(mux, "/api").Handle("GET /users", h)
//	  → registers "GET /api/users"
func Prefix(mux ServeMux, prefix string) Mux {
	return &prefixMux{mux, prefix}
}

type prefixMux struct {
	mux    ServeMux
	prefix string
}

func (m *prefixMux) SubMux() ServeMux {
	return m.mux
}

func (m *prefixMux) prefixPattern(pattern string) string {
	method, path, ok := strings.Cut(pattern, " ")
	if ok {
		return method + " " + m.prefix + path
	}
	return m.prefix + pattern
}

func (m *prefixMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

func (m *prefixMux) Handle(pattern string, handler http.Handler) {
	m.mux.Handle(m.prefixPattern(pattern), handler)
}

func (m *prefixMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.Handle(pattern, http.HandlerFunc(handler))
}

func (m *prefixMux) Route(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	m.Handle(pattern, ApplyMuxErrorHandler(m.mux, handler))
}
