# httpx

Enhanced HTTP utilities for Go, extending `net/http` with middleware support, error handling, and composable mux wrappers.

## Installation

```bash
go get github.com/eriicafes/httpx
```

## Features

- **Error-returning handlers** - Write handlers that return errors
- **Graceful shutdown** - Built-in graceful server shutdown
- **Middleware support** - Apply middlewares to routes
- **Route prefixing** - Group routes under a common prefix
- **Trailing slash normalization** - Automatically handle routes with/without trailing slashes
- **Custom error handling** - Define how errors from handlers are processed
- **Composable** - Chain multiple mux wrappers together

## Subpackages

- [**httperrors**](httperrors/) - Structured HTTP error handling with status codes and details
- [**contextkey**](contextkey/) - Type-safe context key management with generics
- [**session**](session/) - Session-based authentication, cookies, and flash messages

## Quick Start

### Basic Usage

```go
import "github.com/eriicafes/httpx"

mux := httpx.New()

// Regular handler
mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, World!")
})

// Error-returning handler with different error types
mux.Route("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) error {
    id := r.PathValue("id")
    user, err := getUser(id)
    if err != nil {
        // Internal error - logs underlying error, returns user-friendly message
        // Logs: "[ERROR] GET /users/123: sql: no rows in result set"
        // Returns: {"error": "Failed to retrieve user"} with status 500
        return httpx.InternalError(err, "Failed to retrieve user")
    }

    // HTTP error - custom status code and message
    // Returns: {"error": "User email not verified"} with status 403
    if !user.Verified {
        return httperrors.New("User email not verified", http.StatusForbidden)
    }

    // plain error - returns 400 with error message
    // Returns: {"error": "user has been deleted"} with status 400
    if user.Deleted {
        return fmt.Errorf("user has been deleted")
    }

    return httpx.Send(w, user)
})

httpx.ListenAndServe(":8080", mux, nil)
```

### Graceful Shutdown

Start your server with automatic graceful shutdown:

```go
import "github.com/eriicafes/httpx"

mux := httpx.New()
mux.HandleFunc("GET /", handler)

// Simple usage with defaults (30s shutdown timeout)
httpx.ListenAndServe(":8080", mux, nil)

// Custom shutdown timeout
httpx.ListenAndServe(":8080", mux, &httpx.ServerConfig{
    ShutdownTimeout: 60 * time.Second,
})

// Custom server configuration
httpx.ListenAndServe(":8080", mux, &httpx.ServerConfig{
    Server: &http.Server{
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  120 * time.Second,
    },
    ShutdownTimeout: 45 * time.Second,
})
```

The server gracefully shuts down when receiving SIGINT/SIGTERM signals, completing in-flight requests before stopping.

### Middleware

```go
import "github.com/eriicafes/httpx"

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[logging] before %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("[logging] after %s %s", r.Method, r.URL.Path)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("[auth] before")
		next.ServeHTTP(w, r)
		log.Println("[auth] after")
	})
}

// Apply middleware to all routes
mux := httpx.Use(
	http.NewServeMux(),
	loggingMiddleware,
	authMiddleware,
)

mux.HandleFunc("GET /users", handler)
```

Request life cycle:

```text
Request enters
    |
    v
+-----------------------+
| loggingMiddleware     |
|  - before             |
+-----------------------+
            |
            v
        +-----------------------+
        | authMiddleware        |
        |  - before             |
        +-----------------------+
                    |
                    v
              +------------------+
              |   /users handler |
              +------------------+
                    |
                    v
        +-----------------------+
        | authMiddleware        |
        |  - after              |
        +-----------------------+
            |
            v
+-----------------------+
| loggingMiddleware     |
|  - after              |
+-----------------------+
    |
    v
Response exits
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

For each route pattern, two routes are registered: the original pattern and the pattern with exact trailing slash match using `{$}`. Patterns ending with `/` or `{$}` are not duplicated.

### Custom Error Handling

By default, `Route()` and `httpx.HandlerFunc` use the default error handler which:
- Returns JSON responses with `{"error": "error message"}` and status 400 for regular errors
- Returns status 500 for internal errors (wrapped with `InternalError`), logging the underlying error
- Supports `httperrors.HTTPError` for custom status codes and details

You can customize error handling using `Fallback()`:

```go
import (
    "github.com/eriicafes/httpx"
    "github.com/eriicafes/httpx/httperrors"
)

errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
    // Handle structured HTTP errors
    if httpErr, ok := httperrors.Unwrap(err); ok {
        message, statusCode, details := httpErr.HTTPError()
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(map[string]any{
            "error":   message,
            "details": details,
        })
        return
    }
    // Fallback for other errors
    log.Println(r.Method, r.URL.Path, err.Error())
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]any{
        "error": "Internal Server Error",
    })
}

mux := httpx.Fallback(http.NewServeMux(), errorHandler)

mux.Route("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) error {
    user, err := getUser(r.PathValue("id"))
    if err != nil {
        return err
    }
    return httpx.Send(w, user)
})
```

### Composing Wrappers

```go
mux := httpx.New()
mux = httpx.NormalizeTrailingSlash(mux)
mux = httpx.Prefix(mux, "/api/v1")
mux = httpx.Use(mux, loggingMiddleware, authMiddleware)
mux = httpx.Fallback(mux, errorHandler)

// Handler with trailing slash normalization, /api/v1 prefix,
// middleware, and custom error handling
mux.Route("GET /users", func(w http.ResponseWriter, r *http.Request) error {
    // handle request
    return nil
})
```

## Advanced

### Retrieving Error Handlers

Get a mux's error handler or the default if none is configured:

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

### Using HandlerFunc

Convert error-returning functions into `http.Handler`:

```go
// HandlerFunc converts an error-returning function to http.Handler
handler := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
    user, err := getUser(r.PathValue("id"))
    if err != nil {
        // Use InternalError to wrap internal errors
        return httpx.InternalError(err, "Failed to retrieve user")
    }
    return httpx.Send(w, user)
})

http.Handle("GET /users/{id}", handler)
```

**Note:** `HandlerFunc` uses the default error handler which sends JSON responses with status 400 for regular errors and 500 for internal errors. It automatically supports `httperrors.HTTPError` and `InternalError`.

### Converting Existing Handlers

Wrap existing handlers to add httpx features:

```go
// Existing mux
existingMux := http.NewServeMux()
existingMux.HandleFunc("GET /", handlerFunc)

// Add httpx features
mux := httpx.Use(existingMux)
mux = httpx.Fallback(mux, errorHandler)

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
```

**Key points:**
- Implement `ServeHTTP`, `Handle`, `HandleFunc`, and `Route` methods
- Use `SubMux()` to expose the underlying mux for error handler delegation
- Use `httpx.ApplyMuxErrorHandler(mux, handler)` in `Route()` to wrap error-returning handlers

## License

MIT
