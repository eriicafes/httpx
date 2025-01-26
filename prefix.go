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
	ServeMux
	prefix string
}

func (pm *prefixMux) prefixPattern(pattern string) string {
	if strings.Contains(pattern, " ") {
		return strings.Replace(pattern, " ", " "+pm.prefix, 1)
	}
	return pm.prefix + pattern
}

func (pm *prefixMux) Handle(pattern string, handler http.Handler) {
	pm.ServeMux.Handle(pm.prefixPattern(pattern), handler)
}

func (pm *prefixMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	pm.ServeMux.HandleFunc(pm.prefixPattern(pattern), handler)
}

func (pm *prefixMux) Route(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	pm.Handle(pattern, withErrorHandler(pm, handler))
}
