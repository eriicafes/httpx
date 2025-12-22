# httpx

Enhanced HTTP utilities for Go, extending `net/http` with middleware support, error handling, and composable mux wrappers.

## Installation

```bash
go get github.com/eriicafes/httpx
```

## Features

- **Error-returning handlers** - Write handlers that return errors instead of manually calling `http.Error`
- **Middleware support** - Apply middlewares to routes with automatic composition
- **Route prefixing** - Group routes under a common prefix
- **Trailing slash normalization** - Automatically handle routes with/without trailing slashes
- **Custom error handling** - Define how errors from handlers are processed
- **Composable** - Chain multiple mux wrappers together

## Subpackages

- [**httperrors**](httperrors/) - Structured HTTP error handling with status codes
- [**contextkey**](contextkey/) - Type-safe context key management with generics
- [**session**](session/) - Session-based authentication, cookies, and flash messages

## Quick Start

### Basic Usage

```go
import "github.com/eriicafes/httpx"

mux := httpx.New()

// Regular handler
mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
})

// Error-returning handler
mux.Route("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) error {
    id := r.PathValue("id")
    user, err := getUser(id)
    if err != nil {
        return err // Automatically handled
    }
    json.NewEncoder(w).Encode(user)
    return nil
})

http.ListenAndServe(":8080", mux)
```

### Middleware

```go
import "github.com/eriicafes/httpx"

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

// Apply middleware to all routes
mux := httpx.Use(http.NewServeMux(), loggingMiddleware)

mux.HandleFunc("GET /users", handler)
```

### Route Prefixing

```go
import "github.com/eriicafes/httpx"

mux := http.NewServeMux()
api := httpx.Prefix(mux, "/api/v1")

// GET /api/v1/users
api.HandleFunc("GET /users", listUsers)
// GET /api/v1/users/{id}
api.HandleFunc("GET /users/{id}", getUser)

http.ListenAndServe(":8080", mux)
```

### Trailing Slash Normalization

Handle routes with or without trailing slashes:

```go
import "github.com/eriicafes/httpx"

mux := httpx.NormalizeTrailingSlash(http.NewServeMux())

// Both /api/users and /api/users/ will match
mux.HandleFunc("GET /api/users", listUsers)

http.ListenAndServe(":8080", mux)
```

For each route pattern, two routes are registered: the original pattern and the pattern with exact trailing slash match using `{$}`. Patterns already ending with `/` or `{$}` are not duplicated.

### Custom Error Handling

```go
import (
    "github.com/eriicafes/httpx"
    "github.com/eriicafes/httpx/httperrors"
)

errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
    // Handle structured HTTP errors
    if httpErr, ok := httperrors.Unwrap(err); ok {
        message, statusCode, details := httpErr.HTTPError()
        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error":   message,
            "details": details,
        })
        return
    }
    // Fallback for other errors
    http.Error(w, err.Error(), http.StatusInternalServerError)
}

mux := httpx.Fallback(http.NewServeMux(), errorHandler)

mux.Route("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) error {
    user, err := getUser(r.PathValue("id"))
    if err != nil {
        return httperrors.New("User not found", http.StatusNotFound)
    }
    json.NewEncoder(w).Encode(user)
    return nil
})
```

### Composing Wrappers

```go
mux := httpx.New()
mux = httpx.NormalizeTrailingSlash(mux)
mux = httpx.Prefix(mux, "/api/v1")
mux = httpx.Use(mux, loggingMiddleware, authMiddleware)
mux = httpx.Fallback(mux, customErrorHandler)

// Handler with trailing slash normalization, /api/v1 prefix,
// middleware, and custom error handling
mux.Route("GET /users", func(w http.ResponseWriter, r *http.Request) error {
    // handle request
    return nil
})
```

## Advanced

### Layered Error Handlers

Create specialized error handling for different parts of your application:

```go
globalErrorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
    // Handle structured HTTP errors
    if httpErr, ok := httperrors.Unwrap(err); ok {
        message, statusCode, _ := httpErr.HTTPError()
        http.Error(w, message, statusCode)
        return
    }

    // Fallback for other errors
    log.Println(err.Error())
    http.Error(w, "Something went wrong", http.StatusInternalServerError)
}

apiErrorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	// Handle structured HTTP errors
	if httpErr, ok := httperrors.Unwrap(err); ok {
		message, statusCode, _ := httpErr.HTTPError()
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]any{
			"error":   message,
		})
		return
	}

	// Fallback for other errors
    log.Println(err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]any{
		"error": "Something went wrong",
	})
}

mux := httpx.Fallback(http.NewServeMux(), globalErrorHandler)
apiMux := httpx.Fallback(httpx.Prefix(mux, "/api"), apiErrorHandler)

// API routes use apiErrorHandler, other routes use globalErrorHandler
apiMux.Route("GET /users", handler)
mux.Route("GET /health", handler)
```

### Retrieving Error Handlers

Get a mux's error handler to delegate to parent handlers:

```go
parentHandler := httpx.MuxErrorHandler(parentMux)

childErrorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
    if errors.Is(err, ErrUnauthorized) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // Delegate to parent
    parentHandler.HandleError(w, r, err)
}

childMux := httpx.Fallback(parentMux, childErrorHandler)
```

### Wrapping Error Handlers

Get an `http.Handler` that automatically applies the mux's error handler:

```go
handler := func(w http.ResponseWriter, r *http.Request) error {
    return nil
}

// Returns http.Handler with automatic error handling
mux.Handle("GET /custom", httpx.ApplyMuxErrorHandler(mux, handler))
```

### Converting Existing Handlers

Wrap existing handlers to add httpx features:

```go
// Existing mux
existingMux := http.NewServeMux()
existingMux.HandleFunc("GET /legacy", legacyHandler)

// Add httpx features
mux := httpx.Use(existingMux, loggingMiddleware)
mux = httpx.Fallback(mux, customErrorHandler)

// Use error-returning handlers
mux.Route("GET /api/data", func(w http.ResponseWriter, r *http.Request) error {
    return nil
})

// Original routes still work
http.ListenAndServe(":8080", mux)
```

### Implementing Custom Mux

Create custom mux wrappers by implementing the `Mux` interface:

```go
type LoggingMux struct {
    mux httpx.ServeMux
}

// Expose underlying mux for error handler delegation
func (m *LoggingMux) SubMux() httpx.ServeMux {
    return m.mux
}

func (m *LoggingMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    m.mux.ServeHTTP(w, r)
}

func (m *LoggingMux) Handle(pattern string, handler http.Handler) {
    log.Printf("Registering route: %s", pattern)
    m.mux.Handle(pattern, handler)
}

func (m *LoggingMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
    log.Printf("Registering route: %s", pattern)
    m.mux.HandleFunc(pattern, handler)
}

func (m *LoggingMux) Route(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
    m.Handle(pattern, httpx.ApplyMuxErrorHandler(m, handler))
}

// Usage
func NewLoggingMux(mux httpx.ServeMux) httpx.Mux {
    return &LoggingMux{mux: mux}
}

mux := NewLoggingMux(http.NewServeMux())
mux = httpx.Fallback(mux, errorHandler)
```

**Key points:**
- Implement `ServeHTTP`, `Handle`, `HandleFunc`, and `Route` methods
- Use `SubMux()` to expose the underlying mux for error handler delegation
- Use `httpx.ApplyMuxErrorHandler(mux, handler)` in `Route()` to wrap error-returning handlers

## License

MIT
