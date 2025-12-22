package httpx

import (
	"net/http"
	"strings"
)

// Prefix returns a Mux which registers routes with prefix.
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
	m.mux.HandleFunc(m.prefixPattern(pattern), handler)
}

func (m *prefixMux) Route(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	m.Handle(pattern, ApplyMuxErrorHandler(m, handler))
}
