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
func Prefix(mux ServeMux, prefix string) Mux {
	return &prefixMux{mux, prefix}
}

// Group registers a group of routes by calling sub with a Prefixed mux.
//
// Prefix concatenates the prefix with each registered pattern and registers
// the resulting pattern on the underlying mux. Any HTTP method specified in
// the pattern (for example "GET /users") is preserved in the resulting pattern.
// No other parsing, normalization, or rewriting is performed.
func Group(mux ServeMux, prefix string, sub func(Mux)) {
	sub(Prefix(mux, prefix))
}

// MuxPrefix returns the cumulative path prefix accumulated by any Prefix
// mux types in the chain. Returns an empty string if no prefix is found.
func MuxPrefix(mux ServeMux) string {
	prefix := ""
	for {
		switch m := mux.(type) {
		case *prefixMux:
			prefix = m.prefix + prefix
			mux = m.mux
		case Mux:
			mux = m.Mux()
		default:
			return prefix
		}
	}
}

type prefixMux struct {
	mux    ServeMux
	prefix string
}

func (m *prefixMux) Mux() ServeMux {
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
	m.Handle(pattern, Handler(m.mux, handler))
}
