# httpx

Extended HTTP utilities for Go, extending `net/http` with middleware support, error handling, and composable mux types.

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
- **Composable** - Chain multiple mux types together

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

### Mount

`Mount` groups routes under a path prefix using a callback. The callback receives a `Prefix`-wrapped mux and may apply additional mux types before registering routes:

```go
mux := httpx.New()

httpx.Mount(mux, "/api/v1", func(mux httpx.Mux) httpx.Mux {
    mux = httpx.Use(mux, authMiddleware)
    mux.Route("GET /users", listUsers)   // GET /api/v1/users
    mux.Route("POST /users", createUser) // POST /api/v1/users
    return mux
})

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

By default, `Route()` and `httpx.HandlerFunc` use the default error handler which sends JSON responses with the following behavior:
- **Regular errors**: Status 400 (Bad Request) with `{"error": "error message"}`
- **Internal errors** (wrapped with `InternalError`): Status 500 (Internal Server Error) with user-friendly message, logs underlying error server-side
- **HTTP errors** (`httperrors.HTTPError`): Custom status codes and details

You can customize error handling using `Fallback()`:

```go
import (
    "github.com/eriicafes/httpx"
    "github.com/eriicafes/httpx/httperrors"
)

errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
    // Handle structured HTTP errors
    if httpErr, ok := httperrors.Unwrap(err); ok {
        message, statusCode, _ := httpErr.HTTPError()
        http.Error(w, message, statusCode)
        return
    }
    // Fallback for other errors
    log.Println("Error:", r.Method, r.URL.Path, err.Error())
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

### Composing Mux

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

### Converting to http.Handler

When you need to pass an `http.Handler` somewhere (like `http.Handle`, middleware, or custom mux implementations), convert error-returning handlers using one of these approaches:

**1. Using `HandlerFunc` (default error handler):**

```go
// Uses the default error handler
handler := httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
    user, err := getUser(r.PathValue("id"))
    if err != nil {
        return httpx.InternalError(err, "Failed to retrieve user")
    }
    return httpx.Send(w, user)
})

http.Handle("GET /users/{id}", handler)
```

**2. Using `Handler` (reuses mux's error handler):**

Useful when you have an existing mux with custom error handling and want to reuse it:

```go
// Create mux with custom error handling
mux := httpx.Fallback(http.NewServeMux(), customErrorHandler)

handler := func(w http.ResponseWriter, r *http.Request) error {
    user, err := getUser(r.PathValue("id"))
    if err != nil {
        return err
    }
    return httpx.Send(w, user)
}

// Reuses mux's custom error handler
http.Handle("GET /users/{id}", httpx.Handler(mux, handler))
```

### Converting Existing Mux

Wrap an existing mux to add httpx features:

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

Create custom mux types by implementing the `Mux` interface:

```go
type LoggingMux struct {
    mux httpx.ServeMux
}

// Expose underlying mux for error handler delegation
func (m *LoggingMux) Mux() httpx.ServeMux {
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
    m.Handle(pattern, httpx.Handler(m, handler))
}

// Usage
func NewLoggingMux(mux httpx.ServeMux) httpx.Mux {
    return &LoggingMux{mux: mux}
}

mux := NewLoggingMux(http.NewServeMux())
```

**Key points:**
- Implement `ServeHTTP`, `Handle`, `HandleFunc`, and `Route` methods
- Use `Mux()` to expose the underlying mux for error handler delegation
- Use `httpx.Handler(mux, handler)` in `Route()` to wrap error-returning handlers

## License

MIT
