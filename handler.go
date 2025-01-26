package httpx

import "net/http"

// HandlerFunc is a handler func that returns an error
type HandlerFunc func(http.ResponseWriter, *http.Request) error

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

// Middleware is a function that accepts a handler and returns another handler.
// Middlewares can decide to call or not to call the handler and they can run before or after the handler.
type Middleware func(handler http.Handler) http.Handler

type HandleError interface {
	HandleError(http.ResponseWriter, *http.Request, error)
}

// ErrorHandler returns the error handler func for mux.
// If mux does not implement the HandleError interface the returned handler func will write a default error response.
func ErrorHandler(mux ServeMux, err error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h, ok := mux.(HandleError); ok {
			h.HandleError(w, r, err)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func withErrorHandler(mux ServeMux, handler func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			ErrorHandler(mux, err)(w, r)
		}
	}
}
