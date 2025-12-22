package httpx

import "net/http"

// Fallback returns a Mux which executes errorHandler when the handler returns a non-nil error.
func Fallback(mux ServeMux, errorHandler ErrorHandlerFunc) Mux {
	return &fallbackMux{mux, errorHandler}
}

type fallbackMux struct {
	mux          ServeMux
	errorHandler func(http.ResponseWriter, *http.Request, error)
}

func (m *fallbackMux) SubMux() ServeMux {
	return m.mux
}

func (m *fallbackMux) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	m.errorHandler(w, r, err)
}

func (m *fallbackMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

func (m *fallbackMux) Handle(pattern string, handler http.Handler) {
	m.mux.Handle(pattern, handler)
}

func (m *fallbackMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.mux.HandleFunc(pattern, handler)
}

func (m *fallbackMux) Route(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	m.Handle(pattern, ApplyMuxErrorHandler(m, handler))
}
