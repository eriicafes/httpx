package httpx

import "net/http"

// Fallback returns a Mux which executes errorHandler when the handler returns a non-nil error.
func Fallback(mux ServeMux, errorHandler ErrorHandlerFunc) Mux {
	return &fallbackMux{mux, errorHandler}
}

type fallbackMux struct {
	ServeMux
	errorHandler func(http.ResponseWriter, *http.Request, error)
}

func (fm *fallbackMux) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	fm.errorHandler(w, r, err)
}

func (fm *fallbackMux) Route(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	fm.Handle(pattern, withErrorHandler(fm, handler))
}
