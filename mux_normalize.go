package httpx

import (
	"net/http"
	"strings"
)

// NormalizeTrailingSlash returns a Mux that registers an exact trailing
// slash variant for each base pattern, so both "/path" and "/path/" match
// the same handler. Patterns already ending with "/" or "/{$}" are unchanged.
//
// By default, Go's mux treats these pattern types as follows:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/api", h)     // base: exact match only
//	  → GET /api        → 200
//	  → GET /api/       → 404
//	  → GET /api/nested → 404
//
//	mux.HandleFunc("/api/", h)    // subtree: matches /api/ and all paths below
//	  → GET /api        → 301 redirect to /api/
//	  → GET /api/       → 200
//	  → GET /api/nested → 200
//
//	mux.HandleFunc("/api/{$}", h) // exact trailing slash: matches /api/ only
//	  → GET /api        → 301 redirect to /api/
//	  → GET /api/       → 200
//	  → GET /api/nested → 404
//
// To support both "/api" and "/api/", you would have to register "/api/{$}"
// manually and still accept that "/api" redirects to "/api/" instead of
// being served directly. NormalizeTrailingSlash solves this by registering
// both "/api" and "/api/{$}":
//
//	mux := httpx.NormalizeTrailingSlash(mux)
//	mux.HandleFunc("/api", h) // registers both /api and /api/{$}
//	  → GET /api        → 200
//	  → GET /api/       → 200
//	  → GET /api/nested → 404
func NormalizeTrailingSlash(mux ServeMux) Mux {
	return &normalizeTrailingSlashMux{mux}
}

type normalizeTrailingSlashMux struct {
	mux ServeMux
}

func (m *normalizeTrailingSlashMux) Mux() ServeMux {
	return m.mux
}

func (m *normalizeTrailingSlashMux) patterns(pattern string) (pattern1, pattern2 string) {
	if strings.HasSuffix(pattern, "/") {
		return pattern, ""
	}
	if strings.HasSuffix(pattern, "/{$}") {
		return pattern, ""
	}
	return pattern, pattern + "/{$}"
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
	m.Handle(pattern, http.HandlerFunc(handler))
}

func (m *normalizeTrailingSlashMux) Route(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	m.Handle(pattern, Handler(m.mux, handler))
}
