package httpx

import (
	"net/http"
	"strings"
)

// NormalizeTrailingSlash returns a Mux which registers routes with normalized trailing slash.
// For each route it registers "/path" and "/path/{$}".
// This serves both the base path and the exact trailing slash path.
func NormalizeTrailingSlash(mux ServeMux) Mux {
	return &normalizeTrailingSlashMux{mux}
}

type normalizeTrailingSlashMux struct {
	mux ServeMux
}

func (m *normalizeTrailingSlashMux) SubMux() ServeMux {
	return m.mux
}

func (m *normalizeTrailingSlashMux) patterns(pattern string) (pattern1, pattern2 string) {
	if strings.HasSuffix(pattern, "/") {
		return pattern, ""
	}
	if strings.HasSuffix(pattern, "{$}") {
		return pattern, ""
	}
	return pattern, pattern + "{$}"
}

func (m *normalizeTrailingSlashMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

func (m *normalizeTrailingSlashMux) Handle(pattern string, handler http.Handler) {
	pattern1, pattern2 := m.patterns(pattern)
	m.mux.Handle(pattern1, handler)
	if pattern2 != "" {
		m.mux.Handle(pattern2, handler)
	}
}

func (m *normalizeTrailingSlashMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	pattern1, pattern2 := m.patterns(pattern)
	m.mux.HandleFunc(pattern1, handler)
	if pattern2 != "" {
		m.mux.HandleFunc(pattern2, handler)
	}
}

func (m *normalizeTrailingSlashMux) Route(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	m.Handle(pattern, ApplyMuxErrorHandler(m, handler))
}
