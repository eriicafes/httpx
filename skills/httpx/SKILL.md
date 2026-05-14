---
name: httpx
description: Use when building or modifying a Go HTTP server with github.com/eriicafes/httpx, including error-returning handlers, mux composition, middleware, route groups and prefixes, graceful shutdown, JSON helpers, and related subpackages like openapi, session, httperrors, and contextkey.
---

# Httpx

## When to Use

Use this skill when the user wants to:

- create or modify an HTTP server built with `github.com/eriicafes/httpx`
- add routes with handlers that return `error`
- compose mux with middleware, prefixes, fallback error handling, or trailing-slash normalization
- wire graceful startup and shutdown with `httpx.ListenAndServe`

Open these references only when needed:

- `references/openapi.md` for OpenAPI document generation and documented routes
- `references/session.md` for session auth, cookies, and flash messages

## Core Guidance

Start with `mux := httpx.New()` when you want an `httpx.Mux` that supports `Route(pattern, handler)` with handlers returning `error`. Use `httpx.Use(mux)` when adapting an existing `*http.ServeMux`.

```go
mux := httpx.New()

mux.Route("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) error {
	user, err := loadUser(r.PathValue("id"))
	if err != nil {
		return httpx.InternalError(err, "Failed to load user")
	}
	return httpx.Send(w, user)
})
```

Use `Route` for application handlers that may fail. Return:

- a plain `error` for a default `400` JSON response
- `httpx.InternalError(err, message)` when the underlying error should be logged but the client should only see a safe message
- a `httperrors` error when you need an explicit status code or structured response metadata

Use `Handle` or `HandleFunc` when you already have standard `net/http` handlers and do not need `error` returns.

## Httperrors

Prefer:

- `httperrors.New(message, status, extras...)` for fresh HTTP errors
- `httperrors.Wrap(err, message, status, extras...)` to preserve an underlying error while changing the client-facing message
- `httperrors.Report(err, status, extras...)` when the original error message is safe to expose
- `httperrors.Unwrap(err)` when you need to detect or extract an `HTTPError` from a wrapped error chain
- `httperrors.Parse(err)` when you need response-ready values like message, status code, and JSON data from any error

Extras can include:

- `httperrors.Code("SOME_CODE")`
- `httperrors.Details{...}` for field validation errors
- `httperrors.Fields{...}` for other top-level metadata

```go
if err := validate(input); err != nil {
	return httperrors.New("Invalid input", http.StatusBadRequest, httperrors.Details{
		"email": "must be valid",
	})
}
```

The default `httpx` error handler unwraps these errors and sends their structured JSON payload automatically.

## JSON Helpers

Use `httpx.Read[T](r)` to decode JSON requests and `httpx.Send` or `httpx.SendStatus` to write JSON responses.

```go
func createUser(w http.ResponseWriter, r *http.Request) error {
	input, err := httpx.Read[CreateUserInput](r)
	if err != nil {
		return err
	}

	user, err := create(input)
	if err != nil {
		return err
	}

	return httpx.SendStatus(w, http.StatusCreated, user)
}
```

## Context Values

Use `github.com/eriicafes/httpx/contextkey` for typed request-scoped values in `context.Context`.

Define package-level keys with explicit types:

```go
var userKey = contextkey.New[*User]("user")
```

Use:

- `key.Set(ctx, value)` to store a value
- `key.Get(ctx)` to read a value with a typed `(value, ok)` result
- `key.Delete(ctx)` to remove a value
- `key.Take(ctx)` to read and remove in one step
- `key.Update(ctx, fn)` to update from the previous value

For pointer-typed keys, a stored `nil` pointer still counts as a present value. That means `Get` can return `(nil, true)` when the key exists but the stored pointer is nil.

This fits naturally with `httpx` middleware:

```go
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := authenticate(r)
		ctx := userKey.Set(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

With `contextkey`, key identity is tied to the key's concrete type and name. Keep the `(type, name)` combination unique for each logical value you store in context.

## Mux Composition

Compose behavior by wrapping a mux. These wrappers preserve the underlying mux chain, so they can be combined freely.

```go
mux := httpx.New()
mux = httpx.NormalizeTrailingSlash(mux)
mux = httpx.Prefix(mux, "/api")
mux = httpx.Use(mux, loggingMiddleware, authMiddleware)
```

Use:

- `httpx.Use(mux, middlewares...)` to apply middleware in registration order
- `httpx.Prefix(mux, "/api/v1")` to prepend a path prefix while preserving method-aware patterns like `"GET /users"`
- `httpx.Group(mux, "/api/v1", func(mux httpx.Mux) { ... })` for grouped route registration
- `httpx.NormalizeTrailingSlash(mux)` when `"/path"` and `"/path/"` should hit the same handler
- `httpx.Fallback(mux, handler)` to override how returned errors become responses

Prefer wrapping once near app setup instead of scattering wrapper creation across handlers.

Important note for `Prefix` and `Group`: prefixing is just string concatenation. It preserves the optional HTTP method part of a pattern, then concatenates `prefix + path`. It does not normalize slashes, trim separators, or do any special path joining.

## Server Setup

Use `httpx.ListenAndServe(addr, handler, config)` when you want built-in graceful shutdown on `SIGINT` or `SIGTERM`.

```go
err := httpx.ListenAndServe(":8080", mux, &httpx.ServerConfig{
	ShutdownTimeout: 45 * time.Second,
})
```

Prefer passing `nil` for `config` unless you need a custom `http.Server` or a non-default shutdown timeout. When `config` is `nil`, `httpx` uses a default `http.Server` and a `30s` shutdown timeout.
